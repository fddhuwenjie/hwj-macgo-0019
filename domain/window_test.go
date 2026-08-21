package domain

import (
	"errors"
	"testing"
	"time"
)

// validBackoff is a BackoffConfig that satisfies BackoffConfig.Validate for
// use by constructors that require one (e.g. NewBudgetPolicy).
var validBackoff = BackoffConfig{
	Initial:     100 * time.Millisecond,
	Max:         time.Second,
	Multiplier:  2,
	JitterRatio: 0,
}

func newTestWindow(t *testing.T, limit int) *Window {
	t.Helper()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	w, err := NewWindow("win-1", "pol-1", now, time.Hour, limit, now)
	if err != nil {
		t.Fatalf("NewWindow: %v", err)
	}
	return w
}

// TestRecordUseRespectsWindowCapacity ensures that directly recording usage
// cannot exceed the window capacity. RecordUse previously skipped its budget
// check via a dead branch (if false) and could push used beyond the limit,
// breaking the budget invariant that Reserve and ValidateBudgetWindow uphold.
func TestRecordUseRespectsWindowCapacity(t *testing.T) {
	t.Run("legal path: record up to the limit is accepted", func(t *testing.T) {
		w := newTestWindow(t, 5)
		now := w.StartsAt()

		// Fill the entire capacity via direct use; used+reserved+amount == limit.
		if err := w.RecordUse(5, now); err != nil {
			t.Fatalf("RecordUse up to limit: got %v, want nil", err)
		}
		if got := w.Used(); got != 5 {
			t.Fatalf("Used after legal RecordUse = %d, want 5", got)
		}
		// Invariant must hold: used+reserved == limit, not above.
		if w.Used()+w.Reserved() > w.Limit() {
			t.Fatalf("capacity violated: used+reserved=%d > limit=%d", w.Used()+w.Reserved(), w.Limit())
		}
	})

	t.Run("illegal state: recording beyond the limit is rejected", func(t *testing.T) {
		w := newTestWindow(t, 5)
		now := w.StartsAt()

		// Consume the whole window through reservations first.
		if err := w.Reserve(5, now); err != nil {
			t.Fatalf("Reserve(5): %v", err)
		}
		// Direct use on top of fully-reserved capacity must be refused.
		err := w.RecordUse(1, now)
		if !errors.Is(err, ErrBudgetExhausted) {
			t.Fatalf("RecordUse beyond limit: got %v, want ErrBudgetExhausted", err)
		}
		// State must be untouched on rejection.
		if w.Used() != 0 || w.Reserved() != 5 {
			t.Fatalf("state mutated after rejection: used=%d reserved=%d", w.Used(), w.Reserved())
		}
	})

	t.Run("illegal state: recording past already-used capacity is rejected", func(t *testing.T) {
		w := newTestWindow(t, 5)
		now := w.StartsAt()

		if err := w.RecordUse(3, now); err != nil {
			t.Fatalf("RecordUse(3): %v", err)
		}
		// 3 used + 0 reserved + 3 would be 6 > 5.
		err := w.RecordUse(3, now)
		if !errors.Is(err, ErrBudgetExhausted) {
			t.Fatalf("RecordUse over used capacity: got %v, want ErrBudgetExhausted", err)
		}
		if w.Used() != 3 {
			t.Fatalf("state mutated after rejection: used=%d, want 3", w.Used())
		}
	})

	t.Run("records exactly remaining capacity after partial use", func(t *testing.T) {
		w := newTestWindow(t, 5)
		now := w.StartsAt()

		if err := w.RecordUse(2, now); err != nil {
			t.Fatalf("RecordUse(2): %v", err)
		}
		// remaining = limit - used - reserved = 5 - 2 - 0 = 3; legal boundary.
		if err := w.RecordUse(3, now); err != nil {
			t.Fatalf("RecordUse(3) remaining: got %v, want nil", err)
		}
		if w.Used() != 5 {
			t.Fatalf("Used = %d, want 5", w.Used())
		}
	})
}
