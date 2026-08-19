package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// FileTaskStore persists scheduler tasks as JSON files.
type FileTaskStore struct {
	mu  sync.Mutex
	dir string
	now func() time.Time
}

func NewFileTaskStore(dir string) *FileTaskStore {
	return &FileTaskStore{dir: dir, now: time.Now}
}

func (s *FileTaskStore) SaveTask(ctx context.Context, task Task) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := task.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("scheduler: create task dir: %w", err)
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = s.now()
	}
	task.UpdatedAt = s.now()
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return fmt.Errorf("scheduler: marshal task: %w", err)
	}
	path := s.taskPath(task.ID)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("scheduler: write task: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("scheduler: rename task: %w", err)
	}
	return nil
}

func (s *FileTaskStore) LoadTask(ctx context.Context, id string) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.taskPath(id))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Task{}, ErrTaskNotFound
		}
		return Task{}, fmt.Errorf("scheduler: read task: %w", err)
	}
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return Task{}, fmt.Errorf("scheduler: decode task: %w", err)
	}
	return task, nil
}

func (s *FileTaskStore) ListRunnable(ctx context.Context, now time.Time) ([]Task, error) {
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
		return nil, fmt.Errorf("scheduler: list tasks: %w", err)
	}
	tasks := make([]Task, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "task-") || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("scheduler: read task file: %w", err)
		}
		var task Task
		if err := json.Unmarshal(data, &task); err != nil {
			return nil, fmt.Errorf("scheduler: decode task file: %w", err)
		}
		if task.Terminated {
			continue
		}
		if !task.NextRun.After(now) {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (s *FileTaskStore) DeleteTask(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.Remove(s.taskPath(id)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("scheduler: delete task: %w", err)
	}
	return nil
}

func (s *FileTaskStore) taskPath(id string) string {
	return filepath.Join(s.dir, fmt.Sprintf("task-%s.json", id))
}
