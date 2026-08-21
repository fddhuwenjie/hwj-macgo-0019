package domain

import (
	"errors"
	"testing"
	"time"
)

func newTestStreak(t *testing.T) *FailureStreak {
	t.Helper()
	now := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)
	s, err := NewFailureStreak("streak-1", "cat-1", "win-1", now)
	if err != nil {
		t.Fatalf("NewFailureStreak: %v", err)
	}
	return s
}

// RecordSuccess must clear the failure count, last-failure time, and window
// binding so a stale streak cannot bleed into later retries. Regression test
// for the window/quota observation where count survived a success.
func TestFailureStreak_RecordSuccessClearsStreak(t *testing.T) {
	t0 := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)
	s := newTestStreak(t)

	// Two consecutive failures build up the streak.
	if _, err := s.RecordFailure(t0); err != nil {
		t.Fatalf("RecordFailure #1: %v", err)
	}
	if _, err := s.RecordFailure(t0.Add(time.Second)); err != nil {
		t.Fatalf("RecordFailure #2: %v", err)
	}
	if got := s.Count(); got != 2 {
		t.Fatalf("Count before success = %d, want 2", got)
	}
	if s.LastFailureAt().IsZero() {
		t.Fatal("LastFailureAt should be set after failures")
	}
	if s.CurrentWindowID() != "win-1" {
		t.Fatalf("CurrentWindowID = %q, want win-1", s.CurrentWindowID())
	}

	// Success must fully reset the streak: count, last-failure time, window.
	if err := s.RecordSuccess(t0.Add(2 * time.Second)); err != nil {
		t.Fatalf("RecordSuccess: %v", err)
	}

	if got := s.Count(); got != 0 {
		t.Errorf("Count after success = %d, want 0 (stale failure count would pollute later retries)", got)
	}
	if !s.LastFailureAt().IsZero() {
		t.Errorf("LastFailureAt after success = %v, want zero", s.LastFailureAt())
	}
	if s.CurrentWindowID() != "" {
		t.Errorf("CurrentWindowID after success = %q, want empty", s.CurrentWindowID())
	}

	// After a clean reset, a single new failure must start from 1, not carry
	// the old count forward — this is exactly the window/quota regression.
	if got, err := s.RecordFailure(t0.Add(3 * time.Second)); err != nil || got != 1 {
		t.Errorf("RecordFailure after reset = (%d, %v), want (1, nil)", got, err)
	}
}

// RecordSuccess must reject an already-clean streak so callers cannot drive
// the aggregate into a no-op / illegal state. The legitimate path (a streak
// that actually has failures) is preserved by the test above.
func TestFailureStreak_RecordSuccessRejectsCleanStreak(t *testing.T) {
	t0 := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)
	s := newTestStreak(t)

	// A freshly created streak has no failures recorded.
	err := s.RecordSuccess(t0)
	if err == nil {
		t.Fatal("RecordSuccess on a clean streak should be rejected, got nil")
	}
	var de *Error
	if !errors.As(err, &de) {
		t.Fatalf("expected *domain.Error, got %T", err)
	}
	if de.Kind != ErrorKindPrecondition {
		t.Errorf("error kind = %q, want %q", de.Kind, ErrorKindPrecondition)
	}
	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("error should wrap ErrInvalidTransition, got %v", err)
	}
}

// ValidateBackoffLimit must no longer trip after a success, confirming the
// stale-count bug is fixed end-to-end at the invariant level.
func TestFailureStreak_BackoffLimitClearedAfterSuccess(t *testing.T) {
	t0 := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)
	s := newTestStreak(t)

	const maxFailures = 3
	for i := 0; i < maxFailures; i++ {
		if _, err := s.RecordFailure(t0.Add(time.Duration(i) * time.Second)); err != nil {
			t.Fatalf("RecordFailure #%d: %v", i+1, err)
		}
	}
	if err := ValidateBackoffLimit(s, maxFailures); err == nil {
		t.Fatal("expected suspension required at the limit, got nil")
	}

	if err := s.RecordSuccess(t0.Add(time.Minute)); err != nil {
		t.Fatalf("RecordSuccess: %v", err)
	}
	// With count cleared, the streak is back under the limit and must not
	// demand suspension — the window/quota observation that motivated the fix.
	if err := ValidateBackoffLimit(s, maxFailures); err != nil {
		t.Errorf("ValidateBackoffLimit after success = %v, want nil", err)
	}
}
