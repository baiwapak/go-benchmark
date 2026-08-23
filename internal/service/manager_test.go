package service

import (
	"errors"
	"testing"
	"time"

	"github.com/yourorg/go-benchmark/internal/config"
	"github.com/yourorg/go-benchmark/internal/dto"
	"github.com/yourorg/go-benchmark/internal/i18n"
	"github.com/yourorg/go-benchmark/internal/service/store"
	"github.com/yourorg/go-benchmark/pkg/apperr"
)

func TestCreateValidation(t *testing.T) {
	_ = i18n.EnsureLoaded()
	dir := t.TempDir()
	fs, err := store.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		DefaultLang:     "zh",
		ReportDir:       dir,
		MaxConcurrency:  50,
		MaxDurationSec:  60,
		MaxTimeoutSec:   30,
		MaxRampSec:      30,
		MaxBodyBytes:    65536,
		MaxInflightJobs: 2,
	}
	m := NewManager(cfg, fs)

	_, err = m.Create(dto.CreateBenchRequest{URL: "https://example.com", Confirm: false})
	if err == nil {
		t.Fatal("expected confirm error")
	}

	_, err = m.Create(dto.CreateBenchRequest{URL: "ftp://x", Confirm: true})
	if ae, ok := err.(*apperr.AppError); !ok || ae.Field != "url" {
		t.Fatalf("url err=%v", err)
	}

	_, err = m.Create(dto.CreateBenchRequest{URL: "http://127.0.0.1", Confirm: true})
	if ae, ok := err.(*apperr.AppError); !ok || ae.Field != "url" {
		t.Fatalf("ssrf err=%v", err)
	}

	_, err = m.Create(dto.CreateBenchRequest{
		URL: "https://example.com", Confirm: true, Concurrency: 999,
	})
	if ae, ok := err.(*apperr.AppError); !ok || ae.Field != "concurrency" {
		t.Fatalf("concurrency err=%v", err)
	}
}

func TestDeleteAndClearReports(t *testing.T) {
	_ = i18n.EnsureLoaded()
	dir := t.TempDir()
	fs, err := store.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		DefaultLang:     "zh",
		ReportDir:       dir,
		MaxConcurrency:  50,
		MaxDurationSec:  60,
		MaxTimeoutSec:   30,
		MaxRampSec:      30,
		MaxBodyBytes:    65536,
		MaxInflightJobs: 2,
	}
	m := NewManager(cfg, fs)
	now := time.Now()
	a := "aaaaaaaaaaaaaaaa"
	b := "bbbbbbbbbbbbbbbb"
	if err := fs.Save(&dto.BenchReport{ID: a, URL: "https://a.example", Status: "done", FinishedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := fs.Save(&dto.BenchReport{ID: b, URL: "https://b.example", Status: "done", FinishedAt: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Get(a); err != nil {
		t.Fatal(err)
	}
	if err := m.DeleteReport(a); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Get(a); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("want not found after delete, got %v", err)
	}
	n, err := m.ClearReports()
	if err != nil || n != 1 {
		t.Fatalf("clear n=%d err=%v", n, err)
	}
	list, err := m.ListReports(10)
	if err != nil || len(list) != 0 {
		t.Fatalf("want empty list, got %d err=%v", len(list), err)
	}
}
