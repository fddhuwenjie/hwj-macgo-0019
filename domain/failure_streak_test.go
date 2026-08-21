package domain

import (
	"errors"
	"testing"
	"time"
)

// fixedNow is a deterministic reference time for all FailureStreak tests.
// Avoid time.Now() so assertions are stable.
var fixedNow = time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)

func newStreakForTest(t *testing.T, windowID string) *FailureStreak {
	t.Helper()
	s, err := NewFailureStreak("streak-1", "cat-1", windowID, fixedNow)
	if err != nil {
		t.Fatalf("NewFailureStreak: %v", err)
	}
	return s
}

// When failures are recorded within a window, the count rises and the
// backoff alarm threshold can be reached. This is the legal escalation path.
func TestFailureStreak_RecordFailure_CountsAndTriggersBackoff(t *testing.T) {
	s := newStreakForTest(t, "win-1")
	const threshold = 3

	for i := 1; i < threshold; i++ {
		n, err := s.RecordFailure(fixedNow)
		if err != nil {
			t.Fatalf("RecordFailure[%d]: %v", i, err)
		}
		if n != i {
			t.Fatalf("RecordFailure[%d]: got count %d, want %d", i, n, i)
		}
		if got := s.Count(); got != i {
			t.Fatalf("Count after %d failures: got %d, want %d", i, got, i)
		}
		// below the threshold: no suspension required
		if err := ValidateBackoffLimit(s, threshold); err != nil {
			t.Fatalf("ValidateBackoffLimit below threshold: unexpected %v", err)
		}
	}

	// reaching the threshold escalates to ErrSuspensionRequired
	if _, err := s.RecordFailure(fixedNow); err != nil {
		t.Fatalf("RecordFailure at threshold: %v", err)
	}
	if err := ValidateBackoffLimit(s, threshold); !errors.Is(err, ErrSuspensionRequired) {
		t.Fatalf("ValidateBackoffLimit at threshold: got %v, want ErrSuspensionRequired", err)
	}
}

// The core fix: rotating to a new window must clear the consecutive failure
// count so the alarm threshold is isolated per window. Before the fix the
// count leaked across windows and ValidateBackoffLimit fired on the old
// window's tally even though the new window had no failures yet.
func TestFailureStreak_RotateToWindow_ClearsConsecutiveFailures(t *testing.T) {
	s := newStreakForTest(t, "win-1")
	const threshold = 3

	// accumulate failures up to (but not reaching) the threshold in win-1
	for i := 0; i < threshold-1; i++ {
		if _, err := s.RecordFailure(fixedNow); err != nil {
			t.Fatalf("RecordFailure[%d]: %v", i, err)
		}
	}
	if got := s.Count(); got != threshold-1 {
		t.Fatalf("Count before rotate: got %d, want %d", got, threshold-1)
	}

	// rotate forward to the next window
	nextNow := fixedNow.Add(time.Hour)
	if err := s.RotateToWindow("win-2", nextNow); err != nil {
		t.Fatalf("RotateToWindow: %v", err)
	}

	// count must be reset so the new window starts from a clean slate
	if got := s.Count(); got != 0 {
		t.Fatalf("Count after rotate: got %d, want 0 (threshold must be isolated per window)", got)
	}
	if s.CurrentWindowID() != "win-2" {
		t.Fatalf("CurrentWindowID: got %q, want win-2", s.CurrentWindowID())
	}
	if !s.Since().Equal(nextNow) {
		t.Fatalf("Since: got %v, want %v", s.Since(), nextNow)
	}
	if !s.LastFailureAt().IsZero() {
		t.Fatalf("LastFailureAt: got %v, want zero", s.LastFailureAt())
	}
	// a clean new window must not be in a suspension-required state
	if err := ValidateBackoffLimit(s, threshold); err != nil {
		t.Fatalf("ValidateBackoffLimit on clean rotated window: unexpected %v", err)
	}

	// the new window must be able to accumulate its own failures independently
	if _, err := s.RecordFailure(nextNow); err != nil {
		t.Fatalf("RecordFailure on new window: %v", err)
	}
	if got := s.Count(); got != 1 {
		t.Fatalf("Count after one failure in new window: got %d, want 1", got)
	}
}

// Rotating to the same window is a no-op that preserves state and returns nil.
func TestFailureStreak_RotateToWindow_SameWindowIsNoop(t *testing.T) {
	s := newStreakForTest(t, "win-1")
	if _, err := s.RecordFailure(fixedNow); err != nil {
		t.Fatalf("RecordFailure: %v", err)
	}
	beforeVersion := s.Version()
	beforeCount := s.Count()

	if err := s.RotateToWindow("win-1", fixedNow.Add(time.Second)); err != nil {
		t.Fatalf("RotateToWindow same window: %v", err)
	}
	if s.Count() != beforeCount {
		t.Fatalf("no-op rotate changed count: got %d, want %d", s.Count(), beforeCount)
	}
	if s.Version() != beforeVersion {
		t.Fatalf("no-op rotate bumped version: got %d, want %d", s.Version(), beforeVersion)
	}
}

// Illegal state must be rejected: rotating with an empty window id is invalid.
func TestFailureStreak_RotateToWindow_RejectsEmptyWindowID(t *testing.T) {
	s := newStreakForTest(t, "win-1")
	err := s.RotateToWindow("", fixedNow)
	if err == nil {
		t.Fatal("RotateToWindow with empty id: expected error, got nil")
	}
	var de *Error
	if !errors.As(err, &de) {
		t.Fatalf("RotateToWindow error type: got %T, want *domain.Error", err)
	}
	if de.Kind != ErrorKindInvalidArgument {
		t.Fatalf("RotateToWindow error kind: got %q, want %q", de.Kind, ErrorKindInvalidArgument)
	}
	if !errors.Is(err, ErrInvalidInvariant) {
		t.Fatalf("RotateToWindow error: got %v, want wraps ErrInvalidInvariant", err)
	}
}

// A nil receiver is rejected rather than panicking.
func TestFailureStreak_RotateToWindow_NilRejected(t *testing.T) {
	var s *FailureStreak
	if err := s.RotateToWindow("win-1", fixedNow); !errors.Is(err, ErrNilArgument) {
		t.Fatalf("nil RotateToWindow: got %v, want ErrNilArgument", err)
	}
}

// RecordSuccess fully clears the streak, matching the rotate semantics so the
// alarm threshold cannot leak through a stale success path.
func TestFailureStreak_RecordSuccess_ClearsState(t *testing.T) {
	s := newStreakForTest(t, "win-1")
	for i := 0; i < 2; i++ {
		if _, err := s.RecordFailure(fixedNow); err != nil {
			t.Fatalf("RecordFailure[%d]: %v", i, err)
		}
	}
	if err := s.RecordSuccess(fixedNow); err != nil {
		t.Fatalf("RecordSuccess: %v", err)
	}
	if got := s.Count(); got != 0 {
		t.Fatalf("Count after success: got %d, want 0", got)
	}
	if s.CurrentWindowID() != "" {
		t.Fatalf("CurrentWindowID after success: got %q, want empty", s.CurrentWindowID())
	}
}
