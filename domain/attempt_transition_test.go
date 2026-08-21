package domain

import (
	"errors"
	"testing"
	"time"
)

func fixedNow() time.Time { return time.Unix(1_700_000_000, 0) }

// A freshly created attempt starts in RESERVED and any attempt to skip the
// READY/RUNNING lifecycle (e.g. marking it SUCCEEDED straight away) must be
// rejected as an invalid transition. Regression for the bug where MarkSucceeded
// returned nil from RESERVED because updateState never validated transitions.
func TestAttemptCannotSkipLifecycle(t *testing.T) {
	a, err := NewAttempt("a1", "cat", "pol", "win", "res", 0, fixedNow())
	if err != nil {
		t.Fatalf("NewAttempt: %v", err)
	}
	if got := a.State(); got != AttemptReserved {
		t.Fatalf("initial state = %s, want %s", got, AttemptReserved)
	}

	// RESERVED -> SUCCEEDED is illegal: must pass through READY and RUNNING.
	err = a.MarkSucceeded(fixedNow(), "done")
	if err == nil {
		t.Fatal("MarkSucceeded from RESERVED: expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("MarkSucceeded from RESERVED: expected ErrInvalidTransition, got %v", err)
	}
	// State must be unchanged after a rejected transition.
	if got := a.State(); got != AttemptReserved {
		t.Fatalf("state after rejected MarkSucceeded = %s, want %s", got, AttemptReserved)
	}

	// RESERVED -> FAILED is likewise illegal (failure requires RUNNING first).
	if err := a.MarkFailed(fixedNow(), "boom", nil); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("MarkFailed from RESERVED: expected ErrInvalidTransition, got %v", err)
	}
	// RESERVED -> BACKOFF is illegal too.
	if err := a.MarkBackoff(fixedNow(), "back", nil); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("MarkBackoff from RESERVED: expected ErrInvalidTransition, got %v", err)
	}
}

// The legal step-by-step lifecycle RESERVED -> READY -> RUNNING -> SUCCEEDED
// must remain usable end to end.
func TestAttemptLegalLifecycle(t *testing.T) {
	a, err := NewAttempt("a1", "cat", "pol", "win", "res", 0, fixedNow())
	if err != nil {
		t.Fatalf("NewAttempt: %v", err)
	}

	if err := a.MarkReady(fixedNow(), "ready"); err != nil {
		t.Fatalf("MarkReady: %v", err)
	}
	if got := a.State(); got != AttemptReady {
		t.Fatalf("state after MarkReady = %s, want %s", got, AttemptReady)
	}

	if err := a.MarkRunning(fixedNow(), "running"); err != nil {
		t.Fatalf("MarkRunning: %v", err)
	}
	if got := a.State(); got != AttemptRunning {
		t.Fatalf("state after MarkRunning = %s, want %s", got, AttemptRunning)
	}

	if err := a.MarkSucceeded(fixedNow(), "done"); err != nil {
		t.Fatalf("MarkSucceeded from RUNNING: %v", err)
	}
	if got := a.State(); got != AttemptSucceeded {
		t.Fatalf("state after MarkSucceeded = %s, want %s", got, AttemptSucceeded)
	}
	if a.Reason() != "done" {
		t.Fatalf("reason = %q, want %q", a.Reason(), "done")
	}

	// SUCCEEDED is terminal: no further transitions are allowed.
	if err := a.MarkFailed(fixedNow(), "late", nil); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("MarkFailed from SUCCEEDED: expected ErrInvalidTransition, got %v", err)
	}
}

// A failed attempt follows FAILED -> BACKOFF -> READY -> RUNNING, exercising the
// backoff branch of the lifecycle.
func TestAttemptFailureBackoffLifecycle(t *testing.T) {
	a, err := NewAttempt("a1", "cat", "pol", "win", "res", 0, fixedNow())
	if err != nil {
		t.Fatalf("NewAttempt: %v", err)
	}
	mustMark(t, a.MarkReady(fixedNow(), "ready"))
	mustMark(t, a.MarkRunning(fixedNow(), "running"))

	next := fixedNow().Add(time.Second)
	if err := a.MarkFailed(fixedNow(), "boom", &next); err != nil {
		t.Fatalf("MarkFailed: %v", err)
	}
	if got := a.State(); got != AttemptFailed {
		t.Fatalf("state = %s, want %s", got, AttemptFailed)
	}
	if _, ok := a.NextAttemptAt(); !ok {
		t.Fatal("NextAttemptAt not set after MarkFailed")
	}

	if err := a.MarkBackoff(fixedNow(), "backoff", &next); err != nil {
		t.Fatalf("MarkBackoff: %v", err)
	}
	if got := a.State(); got != AttemptBackoff {
		t.Fatalf("state = %s, want %s", got, AttemptBackoff)
	}

	mustMark(t, a.MarkReady(fixedNow(), "ready again"))
	mustMark(t, a.MarkRunning(fixedNow(), "running again"))
	if err := a.MarkSucceeded(fixedNow(), "ok"); err != nil {
		t.Fatalf("MarkSucceeded: %v", err)
	}
	if got := a.State(); got != AttemptSucceeded {
		t.Fatalf("state = %s, want %s", got, AttemptSucceeded)
	}
}

// RESERVED and READY may short-circuit to TIMED_OUT or TERMINATED (the only
// early-exit transitions allowed before RUNNING), and both are terminal.
func TestAttemptEarlyExitTransitions(t *testing.T) {
	for _, tc := range []struct {
		name string
		to   AttemptState
		mark func(*Attempt, time.Time, string) error
	}{
		{"MarkTimedOut", AttemptTimedOut, func(a *Attempt, now time.Time, r string) error { return a.MarkTimedOut(now, r) }},
		{"MarkTerminated", AttemptTerminated, func(a *Attempt, now time.Time, r string) error { return a.MarkTerminated(now, r) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, err := NewAttempt("a1", "cat", "pol", "win", "res", 0, fixedNow())
			if err != nil {
				t.Fatalf("NewAttempt: %v", err)
			}
			if err := tc.mark(a, fixedNow(), "early exit"); err != nil {
				t.Fatalf("%s from RESERVED: %v", tc.name, err)
			}
			if got := a.State(); got != tc.to {
				t.Fatalf("state = %s, want %s", got, tc.to)
			}
			// Terminal: no further transitions allowed.
			if err := a.MarkSucceeded(fixedNow(), "late"); !errors.Is(err, ErrInvalidTransition) {
				t.Fatalf("MarkSucceeded from %s: expected ErrInvalidTransition, got %v", tc.to, err)
			}
		})
	}
}

func mustMark(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("mark: %v", err)
	}
}
