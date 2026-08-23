package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/yourorg/go-benchmark/internal/dto"
	"github.com/yourorg/go-benchmark/pkg/apperr"
)

type FileStore struct {
	dir string
	mu  sync.Mutex
}

func New(dir string) (*FileStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &FileStore{dir: dir}, nil
}

func (s *FileStore) path(id string) string {
	return filepath.Join(s.dir, id+".json")
}

func (s *FileStore) Save(rep *dto.BenchReport) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path(rep.ID) + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path(rep.ID))
}

func (s *FileStore) Load(id string) (*dto.BenchReport, error) {
	id, err := normalizeID(id)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadUnlocked(id)
}

func (s *FileStore) loadUnlocked(id string) (*dto.BenchReport, error) {
	b, err := os.ReadFile(s.path(id))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, apperr.ErrNotFound
		}
		return nil, err
	}
	var rep dto.BenchReport
	if err := json.Unmarshal(b, &rep); err != nil {
		return nil, err
	}
	return &rep, nil
}

func (s *FileStore) Exists(id string) bool {
	id, err := normalizeID(id)
	if err != nil {
		return false
	}
	_, err = os.Stat(s.path(id))
	return err == nil
}

func (s *FileStore) Delete(id string) error {
	id, err := normalizeID(id)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	err = os.Remove(s.path(id))
	if err != nil {
		if os.IsNotExist(err) {
			return apperr.ErrNotFound
		}
		return err
	}
	return nil
}

// ClearExcept deletes all report JSON files except those whose IDs are in skip.
func (s *FileStore) ClearExcept(skip map[string]struct{}) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, ent := range entries {
		if ent.IsDir() || filepath.Ext(ent.Name()) != ".json" {
			continue
		}
		id := strings.TrimSuffix(ent.Name(), ".json")
		if _, ok := skip[id]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(s.dir, ent.Name())); err != nil && !os.IsNotExist(err) {
			return n, err
		}
		n++
	}
	return n, nil
}

func (s *FileStore) List(limit int) ([]dto.ReportSummary, error) {
	if limit <= 0 {
		limit = 50
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ReportSummary, 0, len(entries))
	for _, ent := range entries {
		if ent.IsDir() || filepath.Ext(ent.Name()) != ".json" {
			continue
		}
		id := strings.TrimSuffix(ent.Name(), ".json")
		if _, err := normalizeID(id); err != nil {
			continue
		}
		rep, err := s.loadUnlocked(id)
		if err != nil {
			continue
		}
		out = append(out, dto.ReportSummary{
			ID:          rep.ID,
			URL:         rep.URL,
			Method:      rep.Method,
			Status:      rep.Status,
			RPS:         rep.RPS,
			SuccessRate: rep.SuccessRate,
			Concurrency: rep.Concurrency,
			DurationSec: rep.DurationSec,
			StartedAt:   rep.StartedAt,
			FinishedAt:  rep.FinishedAt,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].FinishedAt.After(out[j].FinishedAt)
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func normalizeID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if n := len(id); n < 8 || n > 32 {
		return "", apperr.ErrBadRequest
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
			continue
		}
		return "", apperr.ErrBadRequest
	}
	return id, nil
}
