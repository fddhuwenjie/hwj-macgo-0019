package recovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"retryengine/domain"
)

// Snapshot is a point-in-time copy of the retry budget state.
type Snapshot struct {
	Version         uint64               `json:"version"`
	JournalSequence uint64               `json:"journal_sequence"`
	CreatedAt       time.Time            `json:"created_at"`
	Attempts        []domain.Attempt     `json:"attempts"`
	Windows         []domain.Window      `json:"windows"`
	Reservations    []domain.Reservation `json:"reservations"`
	Results         []domain.Result      `json:"results"`
	BackoffPlans    []domain.BackoffPlan `json:"backoff_plans"`
}

// Snapshotter defines snapshot persistence operations.
type Snapshotter interface {
	WriteSnapshot(ctx context.Context, snapshot Snapshot) error
	ReadSnapshot(ctx context.Context, fileName string) (Snapshot, error)
	ListSnapshots(ctx context.Context) ([]string, error)
}

// FileSnapshotter stores snapshots as JSON files under a directory.
type FileSnapshotter struct {
	mu  sync.Mutex
	dir string
	now func() time.Time
}

func NewFileSnapshotter(dir string) *FileSnapshotter {
	return &FileSnapshotter{dir: dir, now: time.Now}
}

func (s *FileSnapshotter) WriteSnapshot(ctx context.Context, snap Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("recovery: create snapshot dir: %w", err)
	}
	snap.CreatedAt = s.now()
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("recovery: marshal snapshot: %w", err)
	}
	name := fmt.Sprintf("snapshot-%020d-%s.json", snap.Version, snap.CreatedAt.UTC().Format("20060102T150405Z"))
	tmp := filepath.Join(s.dir, name+".tmp")
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("recovery: write snapshot temp: %w", err)
	}
	if err := os.Rename(tmp, filepath.Join(s.dir, name)); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("recovery: rename snapshot: %w", err)
	}
	return nil
}

func (s *FileSnapshotter) ReadSnapshot(ctx context.Context, fileName string) (Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.Contains(fileName, string(filepath.Separator)) || filepath.Base(fileName) != fileName {
		return Snapshot{}, errors.New("recovery: invalid snapshot file name")
	}
	data, err := os.ReadFile(filepath.Join(s.dir, fileName))
	if err != nil {
		return Snapshot{}, fmt.Errorf("recovery: read snapshot: %w", err)
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return Snapshot{}, fmt.Errorf("recovery: decode snapshot: %w", err)
	}
	return snap, nil
}

func (s *FileSnapshotter) ListSnapshots(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("recovery: list snapshots: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "snapshot-") || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}
