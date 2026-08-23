package capacity

import "runtime"

// Host describes resources of the machine running go-benchmark (the load generator).
type Host struct {
	CPUs       int
	MemTotalMB int
	MemAvailMB int
}

// Limits are recommended hard caps for a single deployment.
type Limits struct {
	MaxConcurrency  int
	MaxDurationSec  int
	MaxInflightJobs int
}

// DetectHost reads local CPU and memory information.
func DetectHost() Host {
	h := Host{CPUs: runtime.NumCPU()}
	if h.CPUs < 1 {
		h.CPUs = 1
	}
	total, avail := detectMemoryMB()
	h.MemTotalMB = total
	h.MemAvailMB = avail
	return h
}

// Recommend derives limits for this load-generator process, not the target site.
func Recommend(host Host) Limits {
	cpus := host.CPUs
	if cpus < 1 {
		cpus = 1
	}

	conc := cpus * 12
	if host.MemAvailMB > 0 {
		if cap := host.MemAvailMB / 3; cap < conc {
			conc = cap
		}
	} else if host.MemTotalMB > 0 {
		if cap := host.MemTotalMB / 4; cap < conc {
			conc = cap
		}
	}
	if conc < 10 {
		conc = 10
	}
	if conc > 1000 {
		conc = 1000
	}

	duration := 300
	if host.MemAvailMB > 0 && host.MemAvailMB < 1024 {
		duration = 120
	}

	inflight := cpus / 2
	if inflight < 1 {
		inflight = 1
	}
	if inflight > 5 {
		inflight = 5
	}
	if host.MemAvailMB > 0 && host.MemAvailMB < 2048 && inflight > 2 {
		inflight = 2
	}

	return Limits{
		MaxConcurrency:  conc,
		MaxDurationSec:  duration,
		MaxInflightJobs: inflight,
	}
}

// MemGBForDisplay returns memory in GB for UI (prefers available, else total).
func MemGBForDisplay(host Host) int {
	mb := host.MemAvailMB
	if mb <= 0 {
		mb = host.MemTotalMB
	}
	if mb <= 0 {
		return 0
	}
	gb := mb / 1024
	if gb < 1 {
		return 1
	}
	return gb
}
