package loadtest

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Params struct {
	URL         string
	Method      string
	Body        string
	Concurrency int
	Ramp        time.Duration
	Duration    time.Duration
	Timeout     time.Duration
	Headers     map[string]string
}

type Sample struct {
	Latency time.Duration
	Status  int
	Bytes   int64
	ErrKind string // empty if ok; timeout|dns|connect|tls|other
	Success bool
}

type Engine struct {
	onSample func(Sample)
}

func New(onSample func(Sample)) *Engine {
	return &Engine{onSample: onSample}
}

func (e *Engine) Run(ctx context.Context, p Params) error {
	if p.Concurrency < 1 {
		p.Concurrency = 1
	}
	if p.Method == "" {
		p.Method = http.MethodGet
	}
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        p.Concurrency * 2,
		MaxIdleConnsPerHost: p.Concurrency * 2,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   p.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	deadline := time.Now().Add(p.Duration)
	var wg sync.WaitGroup
	var stopped atomic.Bool
	started := time.Now()
	var allowedWorkers atomic.Int32
	if p.Ramp <= 0 {
		allowedWorkers.Store(int32(p.Concurrency))
	}

	go func() {
		<-ctx.Done()
		stopped.Store(true)
	}()

	if p.Ramp > 0 {
		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					elapsed := time.Since(started)
					if elapsed >= p.Ramp {
						allowedWorkers.Store(int32(p.Concurrency))
						return
					}
					n := int(float64(p.Concurrency) * elapsed.Seconds() / p.Ramp.Seconds())
					if n < 1 {
						n = 1
					}
					if n > p.Concurrency {
						n = p.Concurrency
					}
					allowedWorkers.Store(int32(n))
				}
			}
		}()
	}

	for i := 0; i < p.Concurrency; i++ {
		workerID := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				if stopped.Load() || time.Now().After(deadline) {
					return
				}
				if ctx.Err() != nil {
					return
				}
				if p.Ramp > 0 && workerID >= int(allowedWorkers.Load()) {
					time.Sleep(20 * time.Millisecond)
					continue
				}
				sample := e.doOne(ctx, client, p)
				if e.onSample != nil {
					e.onSample(sample)
				}
			}
		}()
	}
	wg.Wait()
	return nil
}

func (e *Engine) doOne(ctx context.Context, client *http.Client, p Params) Sample {
	var body io.Reader
	if p.Body != "" {
		body = strings.NewReader(p.Body)
	}
	req, err := http.NewRequestWithContext(ctx, p.Method, p.URL, body)
	if err != nil {
		return Sample{ErrKind: "other", Success: false}
	}
	req.Header.Set("User-Agent", "go-benchmark/1.0")
	hasCT := false
	for k, v := range p.Headers {
		req.Header.Set(k, v)
		if strings.EqualFold(k, "Content-Type") {
			hasCT = true
		}
	}
	if p.Body != "" && !hasCT {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	resp, err := client.Do(req)
	lat := time.Since(start)
	if err != nil {
		return Sample{
			Latency: lat,
			ErrKind: classifyErr(err),
			Success: false,
		}
	}
	n, _ := io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	ok := resp.StatusCode >= 200 && resp.StatusCode < 400
	return Sample{
		Latency: lat,
		Status:  resp.StatusCode,
		Bytes:   n,
		Success: ok,
	}
}

func classifyErr(err error) string {
	if err == nil {
		return ""
	}
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return "timeout"
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "no such host"), strings.Contains(msg, "dns"):
		return "dns"
	case strings.Contains(msg, "tls"), strings.Contains(msg, "certificate"), strings.Contains(msg, "x509"):
		return "tls"
	case strings.Contains(msg, "connection refused"), strings.Contains(msg, "connect"),
		strings.Contains(msg, "network is unreachable"), strings.Contains(msg, "i/o timeout"):
		return "connect"
	case strings.Contains(msg, "context canceled"), strings.Contains(msg, "context deadline"):
		return "timeout"
	default:
		return "other"
	}
}
