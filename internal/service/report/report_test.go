package report

import (
	"testing"

	"github.com/yourorg/go-benchmark/internal/dto"
	"github.com/yourorg/go-benchmark/internal/i18n"
)

func TestPercentile(t *testing.T) {
	vals := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if got := Percentile(vals, 50); got < 5.4 || got > 5.6 {
		t.Fatalf("p50=%v", got)
	}
	if got := Percentile(vals, 90); got < 9.0 || got > 9.2 {
		t.Fatalf("p90=%v", got)
	}
	if Percentile(nil, 50) != 0 {
		t.Fatal("empty")
	}
}

func TestBuildSuggestionsLowSuccess(t *testing.T) {
	_ = i18n.EnsureLoaded()
	rep := &dto.BenchReport{
		Total:       100,
		Success:     80,
		SuccessRate: 80,
		Latency:     dto.LatencyStats{P50Ms: 10, P99Ms: 20, AvgMs: 12},
	}
	sugs := BuildSuggestions("zh", rep)
	found := false
	for _, s := range sugs {
		if s.Code == "low_success" {
			found = true
			if s.Title == "" || s.Detail == "" {
				t.Fatal("empty suggestion text")
			}
		}
	}
	if !found {
		t.Fatalf("expected low_success, got %#v", sugs)
	}
}

func TestBuildSuggestionsHealthy(t *testing.T) {
	_ = i18n.EnsureLoaded()
	rep := &dto.BenchReport{
		Total:       100,
		Success:     100,
		SuccessRate: 100,
		Latency:     dto.LatencyStats{P50Ms: 20, P99Ms: 40, AvgMs: 25},
	}
	sugs := BuildSuggestions("en", rep)
	if len(sugs) != 1 || sugs[0].Code != "healthy" {
		t.Fatalf("got %#v", sugs)
	}
}
