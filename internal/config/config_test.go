package config

import (
	"os"
	"testing"
)

func TestValidateLimits(t *testing.T) {
	cfg := &Config{
		DefaultLang:     "zh",
		ReportDir:       "data/reports",
		MaxConcurrency:  200,
		MaxDurationSec:  300,
		MaxTimeoutSec:   60,
		MaxInflightJobs: 3,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	cfg.MaxConcurrency = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadDefaults(t *testing.T) {
	_ = os.Unsetenv("SERVER_ADDR")
	_ = os.Unsetenv("DEFAULT_LANG")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerAddr != ":8000" {
		t.Fatalf("addr=%s", cfg.ServerAddr)
	}
	if cfg.DefaultLang != "zh" {
		t.Fatalf("lang=%s", cfg.DefaultLang)
	}
}
