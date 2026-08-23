package report

import (
	"math"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/yourorg/go-benchmark/internal/dto"
	"github.com/yourorg/go-benchmark/internal/i18n"
	"github.com/yourorg/go-benchmark/internal/service/loadtest"
)

type secondBucket struct {
	count   int64
	success int64
	sumLat  float64
}

// Aggregator collects samples and builds metrics.
type Aggregator struct {
	mu        sync.Mutex
	latencies []float64 // ms
	buckets   map[int]*secondBucket
	total     int64
	success   int64
	failed    int64
	bytes     int64
	status    map[int]int64
	errors    map[string]int64
	sumLatMs  float64
	started   time.Time
}

func NewAggregator() *Aggregator {
	return &Aggregator{
		buckets: map[int]*secondBucket{},
		status:  map[int]int64{},
		errors:  map[string]int64{},
		started: time.Now(),
	}
}

func (a *Aggregator) Add(s loadtest.Sample) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.total++
	ms := float64(s.Latency.Microseconds()) / 1000.0
	a.latencies = append(a.latencies, ms)
	a.sumLatMs += ms
	a.bytes += s.Bytes

	sec := int(time.Since(a.started).Seconds())
	b := a.buckets[sec]
	if b == nil {
		b = &secondBucket{}
		a.buckets[sec] = b
	}
	b.count++
	b.sumLat += ms
	if s.Success {
		a.success++
		b.success++
	} else {
		a.failed++
	}
	if s.Status > 0 {
		a.status[s.Status]++
	}
	if s.ErrKind != "" {
		a.errors[s.ErrKind]++
	}
}

func (a *Aggregator) Snapshot(jobID, status string, planned time.Duration) dto.ProgressSnapshot {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.snapshotLocked(jobID, status, planned)
}

func (a *Aggregator) snapshotLocked(jobID, status string, planned time.Duration) dto.ProgressSnapshot {
	elapsed := time.Since(a.started).Seconds()
	if elapsed <= 0 {
		elapsed = 0.001
	}
	var avg float64
	if a.total > 0 {
		avg = a.sumLatMs / float64(a.total)
	}
	var rate float64
	if a.total > 0 {
		rate = float64(a.success) / float64(a.total) * 100
	}
	pct := elapsed / planned.Seconds() * 100
	if pct > 100 {
		pct = 100
	}
	if status == "done" || status == "stopped" {
		pct = 100
	}
	return dto.ProgressSnapshot{
		JobID:        jobID,
		Status:       status,
		ElapsedSec:   round2(elapsed),
		ProgressPct:  round2(pct),
		Total:        a.total,
		Success:      a.success,
		Failed:       a.failed,
		RPS:          round2(float64(a.total) / elapsed),
		SuccessRate:  round2(rate),
		AvgLatencyMs: round2(avg),
		Series:       a.seriesLocked(),
	}
}

func (a *Aggregator) seriesLocked() []dto.TimeSeriesPoint {
	if len(a.buckets) == 0 {
		return nil
	}
	secs := make([]int, 0, len(a.buckets))
	for sec := range a.buckets {
		secs = append(secs, sec)
	}
	sort.Ints(secs)
	out := make([]dto.TimeSeriesPoint, 0, len(secs))
	for _, sec := range secs {
		b := a.buckets[sec]
		pt := dto.TimeSeriesPoint{Second: sec}
		if b.count > 0 {
			pt.RPS = round2(float64(b.count))
			pt.AvgLatencyMs = round2(b.sumLat / float64(b.count))
			pt.SuccessRate = round2(float64(b.success) / float64(b.count) * 100)
		}
		out = append(out, pt)
	}
	return out
}

func (a *Aggregator) BuildReport(meta ReportMeta) dto.BenchReport {
	a.mu.Lock()
	defer a.mu.Unlock()
	elapsed := meta.FinishedAt.Sub(meta.StartedAt).Seconds()
	if elapsed <= 0 {
		elapsed = 0.001
	}
	lat := percentileStats(a.latencies, a.sumLatMs)
	var rate float64
	if a.total > 0 {
		rate = float64(a.success) / float64(a.total) * 100
	}
	statusHist := map[string]int64{}
	for code, n := range a.status {
		statusHist[strconv.Itoa(code)] = n
	}
	errHist := map[string]int64{}
	for k, n := range a.errors {
		errHist[k] = n
	}
	rep := dto.BenchReport{
		ID:              meta.ID,
		Lang:            meta.Lang,
		Status:          meta.Status,
		URL:             meta.URL,
		Method:          meta.Method,
		Concurrency:     meta.Concurrency,
		RampSec:         meta.RampSec,
		DurationSec:     meta.DurationSec,
		TimeoutSec:      meta.TimeoutSec,
		Headers:         meta.Headers,
		StartedAt:       meta.StartedAt,
		FinishedAt:      meta.FinishedAt,
		ElapsedSec:      round2(elapsed),
		Total:           a.total,
		Success:         a.success,
		Failed:          a.failed,
		SuccessRate:     round2(rate),
		RPS:             round2(float64(a.total) / elapsed),
		BytesTotal:      a.bytes,
		Latency:         lat,
		Series:          a.seriesLocked(),
		StatusHistogram: statusHist,
		ErrorHistogram:  errHist,
		ErrorMessage:    meta.ErrorMessage,
	}
	rep.Suggestions = BuildSuggestions(meta.Lang, &rep)
	return rep
}

type ReportMeta struct {
	ID           string
	Lang         string
	Status       string
	URL          string
	Method       string
	Concurrency  int
	RampSec      int
	DurationSec  int
	TimeoutSec   int
	Headers      map[string]string
	StartedAt    time.Time
	FinishedAt   time.Time
	ErrorMessage string
}

func percentileStats(vals []float64, sum float64) dto.LatencyStats {
	if len(vals) == 0 {
		return dto.LatencyStats{}
	}
	cp := append([]float64(nil), vals...)
	sort.Float64s(cp)
	avg := sum / float64(len(cp))
	return dto.LatencyStats{
		MinMs: round2(cp[0]),
		MaxMs: round2(cp[len(cp)-1]),
		AvgMs: round2(avg),
		P50Ms: round2(percentile(cp, 50)),
		P90Ms: round2(percentile(cp, 90)),
		P95Ms: round2(percentile(cp, 95)),
		P99Ms: round2(percentile(cp, 99)),
	}
}

func Percentile(sorted []float64, p float64) float64 {
	return percentile(sorted, p)
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	rank := (p / 100) * float64(len(sorted)-1)
	lo := int(math.Floor(rank))
	hi := int(math.Ceil(rank))
	if lo == hi {
		return sorted[lo]
	}
	w := rank - float64(lo)
	return sorted[lo]*(1-w) + sorted[hi]*w
}

func BuildSuggestions(lang string, r *dto.BenchReport) []dto.Suggestion {
	lang = i18n.Normalize(lang)
	var out []dto.Suggestion

	if r.Total > 0 && r.SuccessRate < 95 {
		out = append(out, sug(lang, "critical", "low_success"))
	}

	if r.Latency.P50Ms > 0 && r.Latency.P99Ms/r.Latency.P50Ms > 5 {
		out = append(out, sug(lang, "warn", "tail_latency"))
	}

	var fivexx, throttle, connect int64
	for code, n := range r.StatusHistogram {
		c, _ := strconv.Atoi(code)
		if c >= 500 && c <= 599 {
			fivexx += n
		}
		if c == 429 || c == 503 {
			throttle += n
		}
	}
	for kind, n := range r.ErrorHistogram {
		switch kind {
		case "dns", "connect", "tls":
			connect += n
		}
	}

	if r.Total > 0 && float64(fivexx)/float64(r.Total) > 0.05 {
		out = append(out, sug(lang, "critical", "server_error"))
	}
	if r.Total > 0 && float64(throttle)/float64(r.Total) > 0.05 {
		out = append(out, sug(lang, "warn", "throttled"))
	}
	if r.Total > 0 && float64(connect)/float64(r.Total) > 0.05 {
		out = append(out, sug(lang, "warn", "connect_error"))
	}
	if r.SuccessRate >= 95 && r.Latency.AvgMs >= 500 {
		out = append(out, sug(lang, "info", "slow_ok"))
	}
	if len(out) == 0 {
		out = append(out, sug(lang, "info", "healthy"))
	}
	return out
}

func sug(lang, severity, code string) dto.Suggestion {
	return dto.Suggestion{
		Severity: severity,
		Code:     code,
		Title:    i18n.T(lang, "suggest."+code+".title"),
		Detail:   i18n.T(lang, "suggest."+code+".detail"),
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
