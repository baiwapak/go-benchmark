package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/yourorg/go-benchmark/internal/config"
	"github.com/yourorg/go-benchmark/internal/dto"
	"github.com/yourorg/go-benchmark/internal/i18n"
	"github.com/yourorg/go-benchmark/internal/security"
	"github.com/yourorg/go-benchmark/internal/service/loadtest"
	"github.com/yourorg/go-benchmark/internal/service/report"
	"github.com/yourorg/go-benchmark/internal/service/store"
	"github.com/yourorg/go-benchmark/pkg/apperr"
)

type JobStatus string

const (
	StatusPending  JobStatus = "pending"
	StatusRunning  JobStatus = "running"
	StatusDone     JobStatus = "done"
	StatusStopped  JobStatus = "stopped"
	StatusFailed   JobStatus = "failed"
)

type Job struct {
	ID          string
	Lang        string
	Status      JobStatus
	Params      loadtest.Params
	DurationSec int
	TimeoutSec  int
	RampSec     int
	StartedAt   time.Time
	FinishedAt  time.Time
	ErrorMsg    string
	Agg         *report.Aggregator
	Report      *dto.BenchReport

	cancel context.CancelFunc
	mu     sync.RWMutex
	subs   map[chan dto.ProgressSnapshot]struct{}
}

type Manager struct {
	cfg   *config.Config
	store *store.FileStore

	mu   sync.Mutex
	jobs map[string]*Job
}

func NewManager(cfg *config.Config, fs *store.FileStore) *Manager {
	return &Manager{
		cfg:   cfg,
		store: fs,
		jobs:  map[string]*Job{},
	}
}

func (m *Manager) Create(req dto.CreateBenchRequest) (*Job, error) {
	lang := i18n.Normalize(req.Lang)
	if !req.Confirm {
		return nil, apperr.Validation("confirm", i18n.T(lang, "err.confirm_required"))
	}
	rawURL := strings.TrimSpace(req.URL)
	if rawURL == "" {
		return nil, apperr.Validation("url", i18n.T(lang, "err.url_required"))
	}
	u, err := url.ParseRequestURI(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, apperr.Validation("url", i18n.T(lang, "err.url_invalid"))
	}
	if !security.ValidateTargetURL(rawURL) {
		return nil, apperr.Validation("url", i18n.T(lang, "err.url_ssrf"))
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodGet
	}
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead:
	default:
		return nil, apperr.Validation("method", i18n.T(lang, "err.method"))
	}

	concurrency := req.Concurrency
	if concurrency <= 0 {
		concurrency = 10
	}
	if concurrency > m.cfg.MaxConcurrency {
		return nil, apperr.Validation("concurrency", i18n.T(lang, "err.concurrency"))
	}

	durationSec := req.DurationSec
	if durationSec <= 0 {
		durationSec = 10
	}
	if durationSec > m.cfg.MaxDurationSec {
		return nil, apperr.Validation("duration_sec", i18n.T(lang, "err.duration"))
	}

	timeoutSec := req.TimeoutSec
	if timeoutSec <= 0 {
		timeoutSec = 10
	}
	if timeoutSec > m.cfg.MaxTimeoutSec {
		return nil, apperr.Validation("timeout_sec", i18n.T(lang, "err.timeout"))
	}

	rampSec := req.RampSec
	if rampSec < 0 {
		rampSec = 0
	}
	if rampSec > m.cfg.MaxRampSec {
		return nil, apperr.Validation("ramp_sec", i18n.T(lang, "err.ramp"))
	}
	if rampSec > durationSec {
		return nil, apperr.Validation("ramp_sec", i18n.T(lang, "err.ramp_duration"))
	}

	bodyText := strings.TrimSpace(req.BodyText)
	if len(bodyText) > m.cfg.MaxBodyBytes {
		return nil, apperr.Validation("body_text", i18n.T(lang, "err.body"))
	}
	switch method {
	case http.MethodGet, http.MethodHead:
		bodyText = ""
	}

	headers := map[string]string{}
	for k, v := range req.Headers {
		headers[k] = v
	}
	if strings.TrimSpace(req.HeadersText) != "" {
		parsed, err := parseHeadersText(req.HeadersText)
		if err != nil {
			return nil, apperr.Validation("headers_text", i18n.T(lang, "err.headers"))
		}
		for k, v := range parsed {
			headers[k] = v
		}
	}

	m.mu.Lock()
	inflight := 0
	for _, j := range m.jobs {
		j.mu.RLock()
		st := j.Status
		j.mu.RUnlock()
		if st == StatusRunning || st == StatusPending {
			inflight++
		}
	}
	if inflight >= m.cfg.MaxInflightJobs {
		m.mu.Unlock()
		return nil, apperr.New(429, 429, i18n.T(lang, "err.inflight"))
	}

	id := newID()
	ctx, cancel := context.WithCancel(context.Background())
	job := &Job{
		ID:          id,
		Lang:        lang,
		Status:      StatusPending,
		DurationSec: durationSec,
		TimeoutSec:  timeoutSec,
		RampSec:     rampSec,
		Params: loadtest.Params{
			URL:         rawURL,
			Method:      method,
			Body:        bodyText,
			Concurrency: concurrency,
			Ramp:        time.Duration(rampSec) * time.Second,
			Duration:    time.Duration(durationSec) * time.Second,
			Timeout:     time.Duration(timeoutSec) * time.Second,
			Headers:     headers,
		},
		Agg:    report.NewAggregator(),
		cancel: cancel,
		subs:   map[chan dto.ProgressSnapshot]struct{}{},
	}
	m.jobs[id] = job
	m.mu.Unlock()

	go m.run(ctx, job)
	return job, nil
}

func (m *Manager) run(ctx context.Context, job *Job) {
	job.mu.Lock()
	job.Status = StatusRunning
	job.StartedAt = time.Now()
	job.mu.Unlock()
	m.broadcast(job)

	ticker := time.NewTicker(500 * time.Millisecond)
	stopTicker := make(chan struct{})
	var tickWG sync.WaitGroup
	tickWG.Add(1)
	go func() {
		defer tickWG.Done()
		for {
			select {
			case <-ticker.C:
				m.broadcast(job)
			case <-ctx.Done():
				return
			case <-stopTicker:
				return
			}
		}
	}()

	engine := loadtest.New(func(s loadtest.Sample) {
		job.Agg.Add(s)
	})
	err := engine.Run(ctx, job.Params)
	ticker.Stop()
	close(stopTicker)
	tickWG.Wait()

	job.mu.Lock()
	job.FinishedAt = time.Now()
	status := StatusDone
	if ctx.Err() != nil {
		status = StatusStopped
	}
	if err != nil && status != StatusStopped {
		status = StatusFailed
		job.ErrorMsg = err.Error()
	}
	job.Status = status
	meta := report.ReportMeta{
		ID:           job.ID,
		Lang:         job.Lang,
		Status:       string(status),
		URL:          job.Params.URL,
		Method:       job.Params.Method,
		Concurrency:  job.Params.Concurrency,
		RampSec:      job.RampSec,
		DurationSec:  job.DurationSec,
		TimeoutSec:   job.TimeoutSec,
		Headers:      job.Params.Headers,
		StartedAt:    job.StartedAt,
		FinishedAt:   job.FinishedAt,
		ErrorMessage: job.ErrorMsg,
	}
	rep := job.Agg.BuildReport(meta)
	job.Report = &rep
	job.mu.Unlock()

	_ = m.store.Save(&rep)
	m.broadcast(job)
}

func (m *Manager) broadcast(job *Job) {
	snap := m.snapshot(job)
	job.mu.RLock()
	subs := make([]chan dto.ProgressSnapshot, 0, len(job.subs))
	for ch := range job.subs {
		subs = append(subs, ch)
	}
	job.mu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- snap:
		default:
		}
	}
}

func (m *Manager) snapshot(job *Job) dto.ProgressSnapshot {
	job.mu.RLock()
	defer job.mu.RUnlock()
	st := string(job.Status)
	planned := job.Params.Duration
	snap := job.Agg.Snapshot(job.ID, st, planned)
	switch job.Status {
	case StatusDone:
		snap.Message = i18n.T(job.Lang, "bench.done")
	case StatusStopped:
		snap.Message = i18n.T(job.Lang, "bench.stopped")
	case StatusFailed:
		snap.Message = i18n.T(job.Lang, "bench.failed")
	default:
		snap.Message = i18n.T(job.Lang, "bench.running")
	}
	return snap
}

func (m *Manager) Get(id string) (*Job, error) {
	m.mu.Lock()
	job, ok := m.jobs[id]
	m.mu.Unlock()
	if ok {
		return job, nil
	}
	rep, err := m.store.Load(id)
	if err != nil {
		return nil, err
	}
	// Read-only reconstructed job after restart.
	job = &Job{
		ID:          rep.ID,
		Lang:        rep.Lang,
		Status:      JobStatus(rep.Status),
		DurationSec: rep.DurationSec,
		TimeoutSec:  rep.TimeoutSec,
		RampSec:     rep.RampSec,
		StartedAt:   rep.StartedAt,
		FinishedAt:  rep.FinishedAt,
		ErrorMsg:    rep.ErrorMessage,
		Report:      rep,
		Params: loadtest.Params{
			URL:         rep.URL,
			Method:      rep.Method,
			Concurrency: rep.Concurrency,
			Ramp:        time.Duration(rep.RampSec) * time.Second,
			Headers:     rep.Headers,
		},
		Agg:  report.NewAggregator(),
		subs: map[chan dto.ProgressSnapshot]struct{}{},
	}
	m.mu.Lock()
	m.jobs[id] = job
	m.mu.Unlock()
	return job, nil
}

func (m *Manager) Stop(id string) error {
	job, err := m.Get(id)
	if err != nil {
		return err
	}
	job.mu.RLock()
	cancel := job.cancel
	st := job.Status
	job.mu.RUnlock()
	if st != StatusRunning && st != StatusPending {
		return nil
	}
	if cancel != nil {
		cancel()
	}
	return nil
}

func (m *Manager) Subscribe(job *Job) (<-chan dto.ProgressSnapshot, func()) {
	ch := make(chan dto.ProgressSnapshot, 8)
	job.mu.Lock()
	job.subs[ch] = struct{}{}
	job.mu.Unlock()
	// push current
	ch <- m.snapshot(job)
	unsub := func() {
		job.mu.Lock()
		delete(job.subs, ch)
		job.mu.Unlock()
		close(ch)
	}
	return ch, unsub
}

func (m *Manager) Progress(job *Job) dto.ProgressSnapshot {
	return m.snapshot(job)
}

func (m *Manager) ReportOf(job *Job) *dto.BenchReport {
	job.mu.RLock()
	defer job.mu.RUnlock()
	return job.Report
}

func (m *Manager) ListReports(limit int) ([]dto.ReportSummary, error) {
	return m.store.List(limit)
}

func (m *Manager) DeleteReport(id string) error {
	m.mu.Lock()
	job, ok := m.jobs[id]
	m.mu.Unlock()
	if ok {
		job.mu.RLock()
		st := job.Status
		job.mu.RUnlock()
		if st == StatusRunning || st == StatusPending {
			return apperr.ErrConflict
		}
	}
	if err := m.store.Delete(id); err != nil {
		return err
	}
	m.mu.Lock()
	delete(m.jobs, id)
	m.mu.Unlock()
	return nil
}

func (m *Manager) ClearReports() (int, error) {
	skip := map[string]struct{}{}
	m.mu.Lock()
	for id, job := range m.jobs {
		job.mu.RLock()
		st := job.Status
		job.mu.RUnlock()
		if st == StatusRunning || st == StatusPending {
			skip[id] = struct{}{}
		}
	}
	m.mu.Unlock()

	n, err := m.store.ClearExcept(skip)
	if err != nil {
		return n, err
	}

	m.mu.Lock()
	for id := range m.jobs {
		if _, keep := skip[id]; keep {
			continue
		}
		delete(m.jobs, id)
	}
	m.mu.Unlock()
	return n, nil
}

func parseHeadersText(text string) (map[string]string, error) {
	out := map[string]string{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx := strings.Index(line, ":")
		if idx <= 0 {
			return nil, apperr.ErrBadRequest
		}
		k := strings.TrimSpace(line[:idx])
		v := strings.TrimSpace(line[idx+1:])
		if k == "" {
			return nil, apperr.ErrBadRequest
		}
		out[k] = v
	}
	return out, nil
}

func newID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
