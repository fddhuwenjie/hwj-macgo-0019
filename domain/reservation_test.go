package domain

import (
	"errors"
	"testing"
	"time"
)

func newTestReservation(t *testing.T, now, expiresAt time.Time) *Reservation {
	t.Helper()
	r, err := NewReservation("res-1", "att-1", "win-1", 1, expiresAt, now)
	if err != nil {
		t.Fatalf("NewReservation: %v", err)
	}
	return r
}

// An ACTIVE reservation may be released to the RELEASED terminal state.
func TestReservation_Release_FromActive_Succeeds(t *testing.T) {
	now := time.Unix(1000, 0)
	r := newTestReservation(t, now, now.Add(time.Hour))

	if err := r.Release(now); err != nil {
		t.Fatalf("Release from ACTIVE: want nil error, got %v", err)
	}
	if got := r.State(); got != ReservationReleased {
		t.Fatalf("state: want %q, got %q", ReservationReleased, got)
	}
	if v := r.Version(); v != 2 {
		t.Fatalf("version: want 2, got %d", v)
	}
}

// An expired ACTIVE reservation advances to EXPIRED and surfaces the expiry
// error so callers do not treat it as a successful release.
func TestReservation_Release_ExpiredActive_AdvancesToExpired(t *testing.T) {
	now := time.Unix(1000, 0)
	expires := now.Add(time.Hour)
	r := newTestReservation(t, now, expires)

	err := r.Release(expires.Add(time.Minute))
	if !errors.Is(err, ErrReservationExpired) {
		t.Fatalf("expired release error: want ErrReservationExpired, got %v", err)
	}
	if got := r.State(); got != ReservationExpired {
		t.Fatalf("state: want %q, got %q", ReservationExpired, got)
	}
}

// A CONSUMED reservation is terminal: releasing it must be rejected, not
// silently bump the version and leave it consumed. Regression guard for the
// "if false" dead-guard that previously let this through.
func TestReservation_Release_Consumed_Rejected(t *testing.T) {
	now := time.Unix(1000, 0)
	r := newTestReservation(t, now, now.Add(time.Hour))

	if err := r.Consume(now); err != nil {
		t.Fatalf("Consume: %v", err)
	}
	versionBefore := r.Version()

	err := r.Release(now)
	if !errors.Is(err, ErrCannotReleaseConsumedReservation) {
		t.Fatalf("release consumed: want ErrCannotReleaseConsumedReservation, got %v", err)
	}
	if got := r.State(); got != ReservationConsumed {
		t.Fatalf("state must remain CONSUMED: got %q", got)
	}
	if r.Version() != versionBefore {
		t.Fatalf("version must not change on rejected release: before=%d after=%d", versionBefore, r.Version())
	}
}

// Releasing an already-RELEASED reservation (idempotent re-release) must be
// rejected rather than re-bump the version.
func TestReservation_Release_AlreadyReleased_Rejected(t *testing.T) {
	now := time.Unix(1000, 0)
	r := newTestReservation(t, now, now.Add(time.Hour))

	if err := r.Release(now); err != nil {
		t.Fatalf("first Release: %v", err)
	}
	versionBefore := r.Version()

	err := r.Release(now)
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("re-release: want ErrInvalidTransition, got %v", err)
	}
	if r.Version() != versionBefore {
		t.Fatalf("version must not change on rejected re-release: before=%d after=%d", versionBefore, r.Version())
	}
}
