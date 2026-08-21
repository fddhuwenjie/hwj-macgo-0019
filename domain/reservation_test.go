package domain

import (
	"errors"
	"testing"
	"time"
)

func newTestReservation(t *testing.T, now time.Time, ttl time.Duration) *Reservation {
	t.Helper()
	r, err := NewReservation(
		"res-1",
		"attempt-1",
		"window-1",
		1,
		now.Add(ttl),
		now,
	)
	if err != nil {
		t.Fatalf("NewReservation: %v", err)
	}
	return r
}

func TestReservationConsumeActive(t *testing.T) {
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	r := newTestReservation(t, now, time.Minute)

	before := r.Version()
	if err := r.Consume(now); err != nil {
		t.Fatalf("Consume active reservation: %v", err)
	}
	if r.State() != ReservationConsumed {
		t.Fatalf("state = %q, want %q", r.State(), ReservationConsumed)
	}
	if r.Version() != before+1 {
		t.Fatalf("version not bumped: got %d, want %d", r.Version(), before+1)
	}
}

func TestReservationConsumeExpiredRejected(t *testing.T) {
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	r := newTestReservation(t, now, time.Minute)

	// Just before the expiry boundary the reservation is still active and
	// can be consumed exactly once.
	stillActive := now.Add(time.Minute - time.Nanosecond)
	if err := r.Consume(stillActive); err != nil {
		t.Fatalf("Consume active reservation: %v", err)
	}

	// A second consume fails because the reservation is now CONSUMED.
	if err := r.Consume(stillActive); err == nil {
		t.Fatal("consuming an already-consumed reservation should fail")
	}
}

func TestReservationConsumeExpiredDoesNotMarkConsumed(t *testing.T) {
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	r := newTestReservation(t, now, time.Minute)

	// Move past the expiry boundary: now > expiresAt.
	pastExpiry := now.Add(2 * time.Minute)
	err := r.Consume(pastExpiry)
	if err == nil {
		t.Fatal("Consume on expired reservation should be rejected")
	}
	if !errors.Is(err, ErrReservationExpired) {
		t.Fatalf("err = %v, want wrapping %v", err, ErrReservationExpired)
	}
	// The illegal consume must not have mutated state to CONSUMED.
	if r.State() == ReservationConsumed {
		t.Fatalf("expired reservation was marked CONSUMED; state = %q", r.State())
	}
}
