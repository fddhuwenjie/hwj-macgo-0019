package scheduler

import (
	"context"
	"errors"
	"time"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusSucceeded TaskStatus = "succeeded"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// Task is a persistent scheduler task with generation, attempts and next run time.
type Task struct {
	ID          string    `json:"id"`
	Generation  int       `json:"generation"`
	AttemptID   string    `json:"attempt_id"`
	State       string    `json:"state"`
	NextRun     time.Time `json:"next_run"`
	Attempts    int       `json:"attempts"`
	MaxAttempts int       `json:"max_attempts"`
	LastError   string    `json:"last_error,omitempty"`
	Terminated  bool      `json:"terminated"`
	Version     uint64    `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (t Task) Validate() error {
	if t.ID == "" {
		return errors.New("scheduler: task id required")
	}
	if t.AttemptID == "" {
		return errors.New("scheduler: attempt id required")
	}
	if t.MaxAttempts < 0 {
		return errors.New("scheduler: max attempts must be non-negative")
	}
	if t.NextRun.IsZero() {
		return errors.New("scheduler: next run time required")
	}
	return nil
}

// TaskStore is the persistence boundary for scheduler tasks.
type TaskStore interface {
	SaveTask(ctx context.Context, task Task) error
	LoadTask(ctx context.Context, id string) (Task, error)
	ListRunnable(ctx context.Context, now time.Time) ([]Task, error)
	DeleteTask(ctx context.Context, id string) error
}

var ErrTaskNotFound = errors.New("scheduler: task not found")
