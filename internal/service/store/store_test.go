package store

import (
	"errors"
	"testing"
	"time"

	"github.com/yourorg/go-benchmark/internal/dto"
	"github.com/yourorg/go-benchmark/pkg/apperr"
)

func sampleReport(id string) *dto.BenchReport {
	return &dto.BenchReport{
		ID:          id,
		URL:         "https://example.com",
		Method:      "GET",
		Status:      "done",
		RPS:         12.5,
		SuccessRate: 99,
		Concurrency: 10,
		DurationSec: 30,
		FinishedAt:  time.Now(),
	}
}

func TestSaveLoadDelete(t *testing.T) {
	fs, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	id := "aabbccdd00112233"
	if err := fs.Save(sampleReport(id)); err != nil {
		t.Fatal(err)
	}
	rep, err := fs.Load(id)
	if err != nil {
		t.Fatal(err)
	}
	if rep.ID != id {
		t.Fatalf("id=%s", rep.ID)
	}
	if err := fs.Delete(id); err != nil {
		t.Fatal(err)
	}
	if _, err := fs.Load(id); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("want not found, got %v", err)
	}
	if err := fs.Delete(id); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("second delete: %v", err)
	}
}

func TestDeleteRejectsPathTraversal(t *testing.T) {
	fs, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"../secret", "a/b", `a\b`, "", "short"} {
		if err := fs.Delete(id); !errors.Is(err, apperr.ErrBadRequest) {
			t.Fatalf("id %q: want bad request, got %v", id, err)
		}
	}
}

func TestClearExcept(t *testing.T) {
	dir := t.TempDir()
	fs, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	keep := "aaaaaaaaaaaaaaaa"
	drop := "bbbbbbbbbbbbbbbb"
	if err := fs.Save(sampleReport(keep)); err != nil {
		t.Fatal(err)
	}
	if err := fs.Save(sampleReport(drop)); err != nil {
		t.Fatal(err)
	}
	n, err := fs.ClearExcept(map[string]struct{}{keep: {}})
	if err != nil || n != 1 {
		t.Fatalf("clear except: n=%d err=%v", n, err)
	}
	if !fs.Exists(keep) {
		t.Fatal("kept report missing")
	}
	if fs.Exists(drop) {
		t.Fatal("dropped report still exists")
	}
	n, err = fs.ClearExcept(nil)
	if err != nil || n != 1 {
		t.Fatalf("clear all: n=%d err=%v", n, err)
	}
	list, err := fs.List(10)
	if err != nil || len(list) != 0 {
		t.Fatalf("list after clear: %d err=%v", len(list), err)
	}
}
