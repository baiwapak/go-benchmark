package capacity

import "testing"

func TestRecommendScalesWithCPU(t *testing.T) {
	small := Recommend(Host{CPUs: 2, MemAvailMB: 8192})
	large := Recommend(Host{CPUs: 16, MemAvailMB: 32768})
	if large.MaxConcurrency <= small.MaxConcurrency {
		t.Fatalf("expected larger host to allow more concurrency: %d vs %d", large.MaxConcurrency, small.MaxConcurrency)
	}
	if small.MaxConcurrency < 10 {
		t.Fatalf("too low: %d", small.MaxConcurrency)
	}
}

func TestRecommendLowMemoryCapsConcurrency(t *testing.T) {
	got := Recommend(Host{CPUs: 32, MemAvailMB: 512})
	if got.MaxConcurrency > 200 {
		t.Fatalf("expected memory cap, got concurrency=%d", got.MaxConcurrency)
	}
	if got.MaxInflightJobs > 2 {
		t.Fatalf("expected low inflight on small memory, got %d", got.MaxInflightJobs)
	}
}
