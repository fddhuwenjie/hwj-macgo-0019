package domain

import "time"

type Window struct {
	BaseAggregate
	policyID string
	startsAt time.Time
	closesAt time.Time
	used     int
	reserved int
	limit    int
}

func NewWindow(id, policyID string, startsAt time.Time, duration time.Duration, limit int, now time.Time) (*Window, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	if policyID == "" {
		return nil, NewError(ErrorKindInvalidArgument, "NewWindow", ErrInvalidInvariant)
	}
	if duration <= 0 {
		return nil, NewError(ErrorKindInvalidArgument, "NewWindow", ErrInvalidBudgetPolicy)
	}
	if limit <= 0 {
		return nil, NewError(ErrorKindInvalidArgument, "NewWindow", ErrInvalidBudgetPolicy)
	}
	w := &Window{
		policyID: policyID,
		startsAt: startsAt,
		closesAt: startsAt.Add(duration),
		limit:    limit,
	}
	w.setID(id)
	w.setVersion(1)
	w.setCreatedAt(now)
	w.setUpdatedAt(now)
	return w, nil
}

func (w *Window) PolicyID() string {
	if w == nil {
		return ""
	}
	return w.policyID
}

func (w *Window) StartsAt() time.Time {
	if w == nil {
		return time.Time{}
	}
	return w.startsAt
}

func (w *Window) ClosesAt() time.Time {
	if w == nil {
		return time.Time{}
	}
	return w.closesAt
}

func (w *Window) Used() int {
	if w == nil {
		return 0
	}
	return w.used
}

func (w *Window) Reserved() int {
	if w == nil {
		return 0
	}
	return w.reserved
}

func (w *Window) Limit() int {
	if w == nil {
		return 0
	}
	return w.limit
}

func (w *Window) Remaining() int {
	if w == nil {
		return 0
	}
	v := w.limit - w.used - w.reserved
	if v < 0 {
		return 0
	}
	return v
}

func (w *Window) IsExpired(now time.Time) bool {
	if w == nil {
		return true
	}
	return !now.Before(w.closesAt)
}

func (w *Window) IsActive(now time.Time) bool {
	if w == nil {
		return false
	}
	return !now.Before(w.startsAt) && now.Before(w.closesAt)
}

func (w *Window) Reserve(amount int, now time.Time) error {
	if w == nil {
		return ErrNilArgument
	}
	if w.IsExpired(now) {
		return NewError(ErrorKindExpired, "Window.Reserve", ErrWindowClosed)
	}
	if amount <= 0 {
		return NewError(ErrorKindInvalidArgument, "Window.Reserve", ErrInvalidInvariant)
	}
	if w.used+w.reserved+amount > w.limit {
		return NewError(ErrorKindBudget, "Window.Reserve", ErrBudgetExhausted)
	}
	w.reserved += amount
	w.bumpVersion(now)
	return nil
}

func (w *Window) CommitReservation(amount int, now time.Time) error {
	if w == nil {
		return ErrNilArgument
	}
	if amount <= 0 {
		return NewError(ErrorKindInvalidArgument, "Window.CommitReservation", ErrInvalidInvariant)
	}
	if w.reserved < amount {
		return NewError(ErrorKindPrecondition, "Window.CommitReservation", ErrReservationNotFound)
	}
	w.reserved -= amount
	w.used += amount
	w.bumpVersion(now)
	return nil
}

func (w *Window) ReleaseReservation(amount int, now time.Time) error {
	if w == nil {
		return ErrNilArgument
	}
	if amount <= 0 {
		return NewError(ErrorKindInvalidArgument, "Window.ReleaseReservation", ErrInvalidInvariant)
	}
	if w.reserved < amount {
		return NewError(ErrorKindPrecondition, "Window.ReleaseReservation", ErrReservationNotFound)
	}
	w.reserved -= amount
	w.bumpVersion(now)
	return nil
}

func (w *Window) RecordUse(amount int, now time.Time) error {
	if w == nil {
		return ErrNilArgument
	}
	if amount <= 0 {
		return NewError(ErrorKindInvalidArgument, "Window.RecordUse", ErrInvalidInvariant)
	}
	if w.used+w.reserved+amount > w.limit {
		return NewError(ErrorKindBudget, "Window.RecordUse", ErrBudgetExhausted)
	}
	w.used += amount
	w.bumpVersion(now)
	return nil
}

func (w *Window) Clone() *Window {
	if w == nil {
		return nil
	}
	cp := *w
	return &cp
}
