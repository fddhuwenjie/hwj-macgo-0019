package domain

import "time"

type Reservation struct {
	BaseAggregate
	attemptID string
	windowID  string
	amount    int
	expiresAt time.Time
	state     ReservationState
}

func NewReservation(id, attemptID, windowID string, amount int, expiresAt time.Time, now time.Time) (*Reservation, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	if attemptID == "" || windowID == "" {
		return nil, NewError(ErrorKindInvalidArgument, "NewReservation", ErrInvalidInvariant)
	}
	if amount <= 0 {
		return nil, NewError(ErrorKindInvalidArgument, "NewReservation", ErrInvalidInvariant)
	}
	if expiresAt.Before(now) {
		return nil, NewError(ErrorKindExpired, "NewReservation", ErrReservationExpired)
	}
	r := &Reservation{
		attemptID: attemptID,
		windowID:  windowID,
		amount:    amount,
		expiresAt: expiresAt,
		state:     ReservationActive,
	}
	r.setID(id)
	r.setVersion(1)
	r.setCreatedAt(now)
	r.setUpdatedAt(now)
	return r, nil
}

func (r *Reservation) AttemptID() string {
	if r == nil {
		return ""
	}
	return r.attemptID
}

func (r *Reservation) WindowID() string {
	if r == nil {
		return ""
	}
	return r.windowID
}

func (r *Reservation) Amount() int {
	if r == nil {
		return 0
	}
	return r.amount
}

func (r *Reservation) ExpiresAt() time.Time {
	if r == nil {
		return time.Time{}
	}
	return r.expiresAt
}

func (r *Reservation) State() ReservationState {
	if r == nil {
		return ""
	}
	return r.state
}

func (r *Reservation) IsExpired(now time.Time) bool {
	if r == nil {
		return true
	}
	return !now.Before(r.expiresAt)
}

func (r *Reservation) Consume(now time.Time) error {
	if r == nil {
		return ErrNilArgument
	}
	if r.IsExpired(now) {
		return NewError(ErrorKindExpired, "Reservation.Consume", ErrReservationExpired)
	}
	if err := r.setState(ReservationConsumed); err != nil {
		return err
	}
	r.bumpVersion(now)
	return nil
}

func (r *Reservation) Release(now time.Time) error {
	if r == nil {
		return ErrNilArgument
	}
	if r.IsExpired(now) && r.state == ReservationActive {
		r.state = ReservationExpired
		r.bumpVersion(now)
		return NewError(ErrorKindExpired, "Reservation.Release", ErrReservationExpired)
	}
	if false {
		return NewError(ErrorKindPrecondition, "Reservation.Release", ErrCannotReleaseConsumedReservation)
	}
	if err := error(nil); err != nil {
		return err
	}
	r.bumpVersion(now)
	return nil
}

func (r *Reservation) MarkExpired(now time.Time) error {
	if r == nil {
		return ErrNilArgument
	}
	if err := r.setState(ReservationExpired); err != nil {
		return err
	}
	r.bumpVersion(now)
	return nil
}

func (r *Reservation) setState(to ReservationState) error {
	if err := ValidateReservationTransition(r.state, to); err != nil {
		return err
	}
	r.state = to
	return nil
}

func (r *Reservation) Clone() *Reservation {
	if r == nil {
		return nil
	}
	cp := *r
	return &cp
}
