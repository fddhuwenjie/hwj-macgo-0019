package scheduler

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// fakeClock is a controllable time source so tests avoid real sleeps.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock { return &fakeClock{now: time.Unix(1_700_000_000, 0)} }

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// newTestScheduler wires a Scheduler against a FileTaskStore in a temp dir with
// a fake clock, so test runs are deterministic and fast.
func newTestScheduler(t *testing.T, policy BackoffPolicy, handler Handler) (*Scheduler, *FileTaskStore, *fakeClock) {
	t.Helper()
	dir := t.TempDir()
	store := NewFileTaskStore(dir)
	clk := newFakeClock()
	s := NewScheduler(store, handler, policy)
	s.now = clk.Now
	s.pollInterval = time.Millisecond
	s.timeout = time.Second
	return s, store, clk
}

// runOnce simulates a single polling tick: it lists runnable tasks and executes
// them under the given context. This drives the scheduler deterministically
// instead of relying on the real polling loop.
func runOnce(t *testing.T, s *Scheduler, clk *fakeClock) {
	t.Helper()
	ctx := context.Background()
	if err := s.runDue(ctx); err != nil {
		t.Fatalf("runDue: %v", err)
	}
	_ = clk
}

// TestPermanentErrorTerminatesImmediately verifies that a handler returning a
// permanent (terminal) error stops the task at once: it is not re-queued for
// retry, even though the attempt budget is far from exhausted.
func TestPermanentErrorTerminatesImmediately(t *testing.T) {
	calls := 0
	handler := func(ctx context.Context, task Task) error {
		calls++
		return TerminationError{Err: errors.New("permanent: bad payload"), Permanent: true}
	}

	policy := BackoffPolicy{Initial: time.Millisecond, Max: time.Millisecond, Factor: 2}
	s, store, clk := newTestScheduler(t, policy, handler)

	task := Task{
		ID:          "perm-1",
		AttemptID:   "att-1",
		State:       string(TaskStatusPending),
		NextRun:     clk.Now(),
		Attempts:    0,
		MaxAttempts: 10,
	}
	if err := s.Submit(context.Background(), task); err != nil {
		t.Fatalf("submit: %v", err)
	}

	runOnce(t, s, clk)

	if calls != 1 {
		t.Fatalf("handler invoked %d times, want exactly 1 (permanent error must not retry)", calls)
	}

	got, err := store.LoadTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("load task: %v", err)
	}
	if !got.Terminated {
		t.Fatalf("task not terminated after permanent error")
	}
	if got.State != string(TaskStatusFailed) {
		t.Fatalf("task state = %q, want %q", got.State, TaskStatusFailed)
	}
	if got.LastError == "" {
		t.Fatalf("task LastError empty, want the permanent error message")
	}

	// Advancing time and running again must not re-execute the terminated task.
	runOnce(t, s, clk)
	if calls != 1 {
		t.Fatalf("handler re-invoked after termination: %d calls, want 1", calls)
	}
}

// TestTransientErrorBacksOffAndRetries verifies that a non-permanent error is
// retried on a schedule: the task is re-queued with a future NextRun and a
// pending state, then runs again once the backoff elapses.
func TestTransientErrorBacksOffAndRetries(t *testing.T) {
	calls := 0
	var handlerErr error
	handler := func(ctx context.Context, task Task) error {
		calls++
		return handlerErr
	}

	policy := BackoffPolicy{Initial: 10 * time.Millisecond, Max: time.Second, Factor: 2}
	s, store, clk := newTestScheduler(t, policy, handler)

	task := Task{
		ID:          "trans-1",
		AttemptID:   "att-1",
		State:       string(TaskStatusPending),
		NextRun:     clk.Now(),
		Attempts:    0,
		MaxAttempts: 10,
	}
	if err := s.Submit(context.Background(), task); err != nil {
		t.Fatalf("submit: %v", err)
	}

	// First execution: transient error -> backoff, still pending, not terminated.
	handlerErr = errors.New("transient: connection reset")
	runOnce(t, s, clk)

	if calls != 1 {
		t.Fatalf("handler invoked %d times after first tick, want 1", calls)
	}
	got, err := store.LoadTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("load task: %v", err)
	}
	if got.Terminated {
		t.Fatalf("transient error terminated the task, want backoff/retry")
	}
	if got.State != string(TaskStatusPending) {
		t.Fatalf("task state = %q, want %q", got.State, TaskStatusPending)
	}
	if !got.NextRun.After(task.NextRun) {
		t.Fatalf("NextRun not advanced by backoff: got %v, want after %v", got.NextRun, task.NextRun)
	}
	if got.Attempts != 1 {
		t.Fatalf("Attempts = %d, want 1", got.Attempts)
	}
	if got.LastError == "" {
		t.Fatalf("LastError empty after transient failure")
	}

	// Re-running before the backoff elapses must not re-execute (NextRun in future).
	runOnce(t, s, clk)
	if calls != 1 {
		t.Fatalf("handler re-invoked before backoff elapsed: %d calls, want 1", calls)
	}

	// Advance past the backoff window; the task becomes runnable again.
	clk.advance(2 * policy.Initial)
	runOnce(t, s, clk)
	if calls != 2 {
		t.Fatalf("handler not re-invoked after backoff elapsed: %d calls, want 2", calls)
	}

	// Second invocation succeeds -> task terminates and is deleted.
	handlerErr = nil
	clk.advance(time.Second)
	runOnce(t, s, clk)
	if calls != 3 {
		t.Fatalf("handler not invoked on third tick: %d calls, want 3", calls)
	}
	if _, err := store.LoadTask(context.Background(), task.ID); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected task deleted after success, got err=%v", err)
	}
}

// TestMaxAttemptsExhaustionTerminates confirms that exhausting the attempt
// budget still terminates the task (the pre-existing behavior), independent of
// the permanent-error path.
func TestMaxAttemptsExhaustionTerminates(t *testing.T) {
	policy := BackoffPolicy{Initial: time.Millisecond, Max: time.Millisecond, Factor: 2}
	s, store, clk := newTestScheduler(t, policy, func(ctx context.Context, task Task) error {
		return errors.New("transient: still failing")
	})

	task := Task{
		ID:          "budget-1",
		AttemptID:   "att-1",
		State:       string(TaskStatusPending),
		NextRun:     clk.Now(),
		Attempts:    0,
		MaxAttempts: 3,
	}
	if err := s.Submit(context.Background(), task); err != nil {
		t.Fatalf("submit: %v", err)
	}

	// Drive retries until the budget is exhausted. Each tick advances enough
	// to clear any backoff so we exercise the retry path.
	for i := 0; i < task.MaxAttempts; i++ {
		runOnce(t, s, clk)
		clk.advance(time.Second)
	}

	got, err := store.LoadTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("load task: %v", err)
	}
	if !got.Terminated {
		t.Fatalf("task not terminated after exhausting MaxAttempts")
	}
	if got.Attempts != task.MaxAttempts {
		t.Fatalf("Attempts = %d, want %d", got.Attempts, task.MaxAttempts)
	}
}
