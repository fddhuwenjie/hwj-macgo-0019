package domain

import (
	"errors"
	"testing"
	"time"
)

func newTestSuspension(t *testing.T, now, until time.Time) *Suspension {
	t.Helper()
	s, err := NewSuspension("susp-1", "cat-1", "att-1", "backoff", until, now)
	if err != nil {
		t.Fatalf("NewSuspension: %v", err)
	}
	return s
}

func TestSuspensionLiftRejectsTerminalState(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	until := now.Add(time.Hour)
	s := newTestSuspension(t, now, until)

	liftAt := now.Add(time.Minute)
	if err := s.Lift(liftAt); err != nil {
		t.Fatalf("first Lift: unexpected error: %v", err)
	}
	firstVersion := s.Version()
	gotLiftedAt, ok := s.LiftedAt()
	if !ok || !gotLiftedAt.Equal(liftAt) {
		t.Fatalf("after first Lift: LiftedAt = %v (ok=%v), want %v", gotLiftedAt, ok, liftAt)
	}

	// The suspension is now terminal: re-lifting must be rejected, must not
	// overwrite the lift time, and must not advance the version.
	err := s.Lift(now.Add(2 * time.Minute))
	if err == nil {
		t.Fatal("second Lift: expected terminal-state rejection, got nil")
	}
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("second Lift: error = %v, want wrapping %v", err, ErrInvalidTransition)
	}
	var derr *Error
	if !errors.As(err, &derr) {
		t.Fatalf("second Lift: error = %v, want *domain.Error", err)
	}
	if derr.Kind != ErrorKindPrecondition {
		t.Fatalf("second Lift: kind = %s, want %s", derr.Kind, ErrorKindPrecondition)
	}
	if derr.Op != "Suspension.Lift" {
		t.Fatalf("second Lift: op = %s, want Suspension.Lift", derr.Op)
	}

	if s.Version() != firstVersion {
		t.Fatalf("version advanced after rejected Lift: got %d, want %d", s.Version(), firstVersion)
	}
	if got, ok := s.LiftedAt(); !ok || !got.Equal(liftAt) {
		t.Fatalf("lift time overwritten after rejected Lift: got %v (ok=%v), want %v", got, ok, liftAt)
	}
	if !s.IsLifted() {
		t.Fatal("IsLifted = false after Lift")
	}
}

func TestSuspensionIsActiveAfterExpiryWithoutLift(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	until := now.Add(time.Hour)
	s := newTestSuspension(t, now, until)

	// Before the deadline: active.
	if !s.IsActive(now) {
		t.Fatalf("IsActive before deadline = false, want true (now=%v until=%v)", now, s.SuspendedUntil())
	}
	if s.IsLifted() {
		t.Fatal("IsLifted = true before any Lift")
	}

	// At the exact deadline: no longer active (the deadline is exclusive).
	if s.IsActive(until) {
		t.Fatalf("IsActive at deadline = true, want false (until=%v)", s.SuspendedUntil())
	}

	// After the deadline: still not active, but purely by expiry — not lifted,
	// so the normal active-status rule applies rather than the terminal lift.
	after := until.Add(time.Minute)
	if s.IsActive(after) {
		t.Fatalf("IsActive after deadline = true, want false")
	}
	if s.IsLifted() {
		t.Fatal("IsLifted = true after expiry without Lift")
	}
	if _, ok := s.LiftedAt(); ok {
		t.Fatal("LiftedAt reported present after expiry without Lift")
	}
}

func TestSuspensionLiftedNotActiveRegardlessOfDeadline(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	until := now.Add(time.Hour)
	s := newTestSuspension(t, now, until)

	if err := s.Lift(now); err != nil {
		t.Fatalf("Lift: %v", err)
	}
	// Lifted early: not active even though the deadline has not arrived.
	if s.IsActive(now) {
		t.Fatal("IsActive after Lift before deadline = true, want false")
	}
	// Lifted: not active even after the deadline would have passed.
	if s.IsActive(until.Add(time.Minute)) {
		t.Fatal("IsActive after Lift past deadline = true, want false")
	}
}
