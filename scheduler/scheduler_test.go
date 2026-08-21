package scheduler

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeClock struct {
	mu  sync.Mutex
	cur time.Time
}

func newFakeClock(start time.Time) *fakeClock {
	return &fakeClock{cur: start}
}

func (f *fakeClock) now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.cur
}

func (f *fakeClock) advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cur = f.cur.Add(d)
}

func testLogger(t *testing.T) *log.Logger {
	return log.New(testWriter{t: t}, "scheduler: ", 0)
}

type testWriter struct{ t *testing.T }

func (w testWriter) Write(p []byte) (int, error) {
	w.t.Logf("%s", p)
	return len(p), nil
}

// newTestScheduler builds a scheduler with a fresh file store and injected clock.
func newTestScheduler(t *testing.T, handler Handler) (*Scheduler, *FileTaskStore, *fakeClock) {
	t.Helper()
	dir := t.TempDir()
	store := NewFileStoreWithDir(dir)
	clock := newFakeClock(time.Unix(1_700_000_000, 0))
	s := &Scheduler{
		store:        store,
		handler:      handler,
		policy:       BackoffPolicy{Initial: 10 * time.Millisecond, Max: 50 * time.Millisecond, Factor: 2},
		pollInterval: 10 * time.Millisecond,
		timeout:      time.Second,
		now:          clock.now,
		stopCh:       make(chan struct{}),
		logger:       testLogger(t),
	}
	store.now = clock.now
	return s, store, clock
}

// NewFileStoreWithDir exposes a file store pointed at a directory for tests.
func NewFileStoreWithDir(dir string) *FileTaskStore {
	return &FileTaskStore{dir: dir, now: time.Now}
}

// TestExecuteRunsExactlyOnceAcrossPolls reproduces the duplicate-execution bug:
// a task whose handler is slower than the poll interval must not be picked up
// by a subsequent poll before its state has been advanced.
func TestExecuteRunsExactlyOnceAcrossPolls(t *testing.T) {
	var starts int32
	handlerStart := make(chan struct{}, 1)
	handlerProceed := make(chan struct{})

	handler := func(ctx context.Context, task Task) error {
		atomic.AddInt32(&starts, 1)
		select {
		case handlerStart <- struct{}{}:
		default:
		}
		select {
		case <-handlerProceed:
		case <-ctx.Done():
		}
		return nil
	}

	s, store, clock := newTestScheduler(t, handler)

	task := Task{
		ID:          "dup-1",
		AttemptID:   "att-1",
		State:       string(TaskStatusPending),
		NextRun:     clock.now(),
		MaxAttempts: 3,
	}
	if err := s.Submit(context.Background(), task); err != nil {
		t.Fatalf("submit: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Wait until the handler is running, then advance the clock past several
	// poll intervals while the handler is still in flight.
	select {
	case <-handlerStart:
	case <-time.After(time.Second):
		t.Fatal("handler never started on first poll")
	}

	for i := 0; i < 5; i++ {
		clock.advance(50 * time.Millisecond)
		time.Sleep(20 * time.Millisecond)
	}

	// With the bug, starts would already be >= 2. With the fix it must still be 1.
	if got := atomic.LoadInt32(&starts); got != 1 {
		close(handlerProceed)
		t.Fatalf("task started %d times while handler was in flight; want exactly 1", got)
	}

	// Let the handler finish; the task should complete and be removed.
	close(handlerProceed)
	if err := waitForTaskRemoved(store, task.ID, 2*time.Second); err != nil {
		t.Fatalf("task not removed after handler completion: %v", err)
	}

	if err := s.Stop(context.Background()); err != nil {
		t.Fatalf("stop: %v", err)
	}

	if got := atomic.LoadInt32(&starts); got != 1 {
		t.Fatalf("task started %d times in total; want exactly 1", got)
	}
}

// TestFastTaskStillCompletes ensures a task that finishes well within one poll
// interval is executed and removed exactly once (the fast-completion path the
// user asked us to preserve).
func TestFastTaskStillCompletes(t *testing.T) {
	var starts int32
	handler := func(ctx context.Context, task Task) error {
		atomic.AddInt32(&starts, 1)
		return nil
	}

	s, store, clock := newTestScheduler(t, handler)
	task := Task{
		ID:          "fast-1",
		AttemptID:   "att-1",
		State:       string(TaskStatusPending),
		NextRun:     clock.now(),
		MaxAttempts: 3,
	}
	if err := s.Submit(context.Background(), task); err != nil {
		t.Fatalf("submit: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() { _ = s.Stop(context.Background()) })

	if err := waitForTaskRemoved(store, task.ID, 2*time.Second); err != nil {
		t.Fatalf("fast task not removed: %v", err)
	}
	if got := atomic.LoadInt32(&starts); got != 1 {
		t.Fatalf("fast task started %d times; want exactly 1", got)
	}
}

// TestStopWaitsForInFlightExecution confirms Stop blocks until any handler
// currently running finishes, rather than racing past it.
func TestStopWaitsForInFlightExecution(t *testing.T) {
	handlerStart := make(chan struct{})
	handlerProceed := make(chan struct{})

	handler := func(ctx context.Context, task Task) error {
		close(handlerStart)
		<-handlerProceed
		return nil
	}

	s, store, clock := newTestScheduler(t, handler)
	task := Task{
		ID:          "stop-1",
		AttemptID:   "att-1",
		State:       string(TaskStatusPending),
		NextRun:     clock.now(),
		MaxAttempts: 3,
	}
	if err := s.Submit(context.Background(), task); err != nil {
		t.Fatalf("submit: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}

	select {
	case <-handlerStart:
	case <-time.After(time.Second):
		t.Fatal("handler never started")
	}

	stopDone := make(chan error, 1)
	go func() { stopDone <- s.Stop(context.Background()) }()

	select {
	case err := <-stopDone:
		t.Fatalf("Stop returned before handler finished: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	close(handlerProceed)
	select {
	case err := <-stopDone:
		if err != nil {
			t.Fatalf("stop: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("stop did not return after handler finished")
	}

	if _, err := store.LoadTask(context.Background(), task.ID); err == nil {
		t.Fatal("task still present after stop completed")
	}
}

func waitForTaskRemoved(store *FileTaskStore, id string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, err := store.LoadTask(context.Background(), id)
		if errors.Is(err, ErrTaskNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)
	}
	return errors.New("task still present")
}
