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
	return &Scheduler{
		store:        store,
		handler:      handler,
		policy:       policy,
		pollInterval: time.Second,
		timeout:      time.Minute,
		now:          time.Now,
		stopCh:       make(chan struct{}),
		logger:       log.Default(),
	}
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
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Scheduler) Submit(ctx context.Context, task Task) error {
	if err := task.Validate(); err != nil {
		return err
	}
	return s.store.SaveTask(ctx, task)
}

// Recover reloads persisted still-legal tasks. The run loop will pick them up.
func (s *Scheduler) Recover(ctx context.Context) error {
	tasks, err := s.store.ListRunnable(ctx, s.now())
	if err != nil {
		return fmt.Errorf("scheduler: recover: %w", err)
	}
	for _, task := range tasks {
		if task.Terminated {
			continue
		}
		// Ensure the task is persisted and visible to the polling loop.
		if task.Version == 0 {
			task.Version = 1
			if err := s.store.SaveTask(ctx, task); err != nil {
				return err
			}
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
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			// runDue blocks until every task it dispatched this round has
			// finished and persisted its new state. Waiting here is what
			// keeps a still-due task from being picked up again by the next
			// poll before its state has advanced.
			if err := s.runDue(ctx); err != nil {
				s.logger.Printf("scheduler: run due: %v", err)
			}
		}
	}
}

// runDue lists due tasks and executes them. It does not return until every
// task dispatched this round has finished executing and written its resulting
// state (succeeded/deleted/retry-scheduled/terminated). This serializes
// polling relative to execution so a slow handler cannot be re-claimed by the
// next poll. A round whose execution outlives the context or stop signal is
// still awaited here so Stop observes its completion.
func (s *Scheduler) runDue(ctx context.Context) error {
	tasks, err := s.store.ListRunnable(ctx, s.now())
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return nil
	}
	var round sync.WaitGroup
	for _, task := range tasks {
		task := task
		round.Add(1)
		go func() {
			defer round.Done()
			if err := s.execute(ctx, task); err != nil {
				s.logger.Printf("scheduler: task %s execute: %v", task.ID, err)
			}
		}()
	}
	round.Wait()
	return nil
}

func (s *Scheduler) execute(ctx context.Context, task Task) error {
	runCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	task.Attempts++
	task.State = string(TaskStatusRunning)
	task.UpdatedAt = s.now()
	if err := s.store.SaveTask(runCtx, task); err != nil {
		return err
	}

	err := s.handler(runCtx, task)
	task.UpdatedAt = s.now()

	if err == nil {
		task.Terminated = true
		task.State = string(TaskStatusSucceeded)
		task.LastError = ""
		if saveErr := s.store.SaveTask(runCtx, task); saveErr != nil {
			return saveErr
		}
		return s.store.DeleteTask(runCtx, task.ID)
	}

	task.LastError = err.Error()
	task.State = string(TaskStatusFailed)
	if IsTerminal(err) || task.Attempts >= task.MaxAttempts {
		task.Terminated = true
		return s.store.SaveTask(runCtx, task)
	}

	delay := s.policy.NextDelay(task.Attempts)
	task.NextRun = s.now().Add(delay)
	task.Generation++
	task.State = string(TaskStatusPending)
	return s.store.SaveTask(runCtx, task)
}
