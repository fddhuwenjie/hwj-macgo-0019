package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// SnapshotStore is an optional local snapshot helper used by recovery.
// It deliberately does not depend on other repository types so it can be
// embedded or reused without forcing a persistence cycle.
type SnapshotStore struct {
	mu  sync.Mutex
	dir string
}

// NewSnapshotStore creates a snapshot store rooted at dir.
func NewSnapshotStore(dir string) (*SnapshotStore, error) {
	if dir == "" {
		return nil, errors.New("repository: snapshot dir is empty")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &SnapshotStore{dir: dir}, nil
}

// List returns snapshot file names sorted by name.
func (s *SnapshotStore) List() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// WriteSnapshot writes a JSON snapshot atomically using a temp file and rename.
func (s *SnapshotStore) WriteSnapshot(name string, value any) error {
	if name == "" {
		return errors.New("repository: snapshot name is empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	tmp := filepath.Join(s.dir, name+".tmp")
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, filepath.Join(s.dir, name)); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}
