package dto

import "time"

type CreateBenchRequest struct {
	URL         string            `json:"url" form:"url"`
	Method      string            `json:"method" form:"method"`
	Concurrency int               `json:"concurrency" form:"concurrency"`
	DurationSec int               `json:"duration_sec" form:"duration_sec"`
	TimeoutSec  int               `json:"timeout_sec" form:"timeout_sec"`
	HeadersText string            `json:"headers_text" form:"headers_text"`
	Headers     map[string]string `json:"headers"`
	Confirm     bool              `json:"confirm" form:"confirm"`
	Lang        string            `json:"lang" form:"lang"`
}

type ProgressSnapshot struct {
	JobID        string  `json:"job_id"`
	Status       string  `json:"status"`
	ElapsedSec   float64 `json:"elapsed_sec"`
	ProgressPct  float64 `json:"progress_pct"`
	Total        int64   `json:"total"`
	Success      int64   `json:"success"`
	Failed       int64   `json:"failed"`
	RPS          float64 `json:"rps"`
	SuccessRate  float64 `json:"success_rate"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	Message      string  `json:"message,omitempty"`
}

type LatencyStats struct {
	MinMs float64 `json:"min_ms"`
	MaxMs float64 `json:"max_ms"`
	AvgMs float64 `json:"avg_ms"`
	P50Ms float64 `json:"p50_ms"`
	P90Ms float64 `json:"p90_ms"`
	P95Ms float64 `json:"p95_ms"`
	P99Ms float64 `json:"p99_ms"`
}

type Suggestion struct {
	Severity string `json:"severity"` // info | warn | critical
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Code     string `json:"code"`
}

type BenchReport struct {
	ID              string             `json:"id"`
	Lang            string             `json:"lang"`
	Status          string             `json:"status"`
	URL             string             `json:"url"`
	Method          string             `json:"method"`
	Concurrency     int                `json:"concurrency"`
	DurationSec     int                `json:"duration_sec"`
	TimeoutSec      int                `json:"timeout_sec"`
	Headers         map[string]string  `json:"headers,omitempty"`
	StartedAt       time.Time          `json:"started_at"`
	FinishedAt      time.Time          `json:"finished_at"`
	ElapsedSec      float64            `json:"elapsed_sec"`
	Total           int64              `json:"total"`
	Success         int64              `json:"success"`
	Failed          int64              `json:"failed"`
	SuccessRate     float64            `json:"success_rate"`
	RPS             float64            `json:"rps"`
	BytesTotal      int64              `json:"bytes_total"`
	Latency         LatencyStats       `json:"latency"`
	StatusHistogram map[string]int64   `json:"status_histogram"`
	ErrorHistogram  map[string]int64   `json:"error_histogram"`
	Suggestions     []Suggestion       `json:"suggestions"`
	ErrorMessage    string             `json:"error_message,omitempty"`
}
