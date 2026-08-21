package store

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	s.mu.Lock()
	defer s.mu.Unlock()
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
	_, err := os.Stat(s.path(id))
	return err == nil
}
