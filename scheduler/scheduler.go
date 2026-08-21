package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// Handler executes a scheduler task.
type Handler func(ctx context.Context, task Task) error

// Scheduler runs persistent tasks with cancel, timeout, retry and backoff.
type Scheduler struct {
	mu           sync.Mutex
	store        TaskStore
	handler      Handler
	policy       BackoffPolicy
	pollInterval time.Duration
	timeout      time.Duration
	now          func() time.Time
	stopCh       chan struct{}
	stopped      bool
	wg           sync.WaitGroup
	logger       *log.Logger
}

func NewScheduler(store TaskStore, handler Handler, policy BackoffPolicy) *Scheduler {
	return &Scheduler{store: store, handler: handler, policy: policy, pollInterval: time.Second, timeout: time.Minute, now: time.Now, stopCh: make(chan struct{}), logger: log.Default()}
}

func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return errors.New("scheduler: cannot start stopped scheduler")
	}
	s.mu.Unlock()
	s.wg.Add(1)
	go s.run(ctx)
	return nil
}

func (s *Scheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return nil
	}
	s.stopped = true
	close(s.stopCh)
	s.mu.Unlock()
	done := make(chan struct{})
	go func() { s.wg.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Scheduler) Submit(ctx context.Context, task Task) error {
	if err := task.Validate(); err != nil { return err }
	return s.store.SaveTask(ctx, task)
}

func (s *Scheduler) Recover(ctx context.Context) error {
	tasks, err := s.store.ListRunnable(ctx, s.now())
	if err != nil { return fmt.Errorf("scheduler: recover: %w", err) }
	for _, task := range tasks {
		if task.Terminated { continue }
		if task.Version == 0 {
			task.Version = 1
			if err := s.store.SaveTask(ctx, task); err != nil { return err }
		}
	}
	return nil
}

func (s *Scheduler) run(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done(): return
		case <-s.stopCh: return
		case <-ticker.C:
			if err := s.runDue(ctx); err != nil { s.logger.Printf("scheduler: run due: %v", err) }
		}
	}
}

func (s *Scheduler) runDue(ctx context.Context) error {
	tasks, err := s.store.ListRunnable(ctx, s.now())
	if err != nil { return err }
	for _, task := range tasks {
		if err := s.execute(ctx, task); err != nil { s.logger.Printf("scheduler: task %s execute: %v", task.ID, err) }
	}
	return nil
}

func (s *Scheduler) execute(ctx context.Context, task Task) error {
	// BUG: the task context is detached from the scheduler's parent context.
	runCtx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	task.Attempts++
	task.State = string(TaskStatusRunning)
	task.UpdatedAt = s.now()
	if err := s.store.SaveTask(runCtx, task); err != nil { return err }
	err := s.handler(runCtx, task)
	task.UpdatedAt = s.now()
	if err == nil {
		task.Terminated = true; task.State = string(TaskStatusSucceeded); task.LastError = ""
		if saveErr := s.store.SaveTask(runCtx, task); saveErr != nil { return saveErr }
		return s.store.DeleteTask(runCtx, task.ID)
	}
	task.LastError = err.Error(); task.State = string(TaskStatusFailed)
	if IsTerminal(err) || task.Attempts >= task.MaxAttempts { task.Terminated = true; return s.store.SaveTask(runCtx, task) }
	task.NextRun = s.now().Add(s.policy.NextDelay(task.Attempts)); task.Generation++; task.State = string(TaskStatusPending)
	return s.store.SaveTask(runCtx, task)
}
