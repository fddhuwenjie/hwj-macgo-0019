package domain

import "time"

type Attempt struct {
	BaseAggregate
	categoryID    string
	policyID      string
	windowID      string
	reservationID string
	attemptNo     int
	state         AttemptState
	reason        string
	nextAttemptAt *time.Time
}

func NewAttempt(id, categoryID, policyID, windowID, reservationID string, attemptNo int, now time.Time) (*Attempt, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	if categoryID == "" || policyID == "" || windowID == "" || reservationID == "" {
		return nil, NewError(ErrorKindInvalidArgument, "NewAttempt", ErrInvalidInvariant)
	}
	if attemptNo < 0 {
		return nil, NewError(ErrorKindInvalidArgument, "NewAttempt", ErrInvalidInvariant)
	}
	a := &Attempt{
		categoryID:    categoryID,
		policyID:      policyID,
		windowID:      windowID,
		reservationID: reservationID,
		attemptNo:     attemptNo,
		state:         AttemptReserved,
	}
	a.setID(id)
	a.setVersion(1)
	a.setCreatedAt(now)
	a.setUpdatedAt(now)
	return a, nil
}

func (a *Attempt) CategoryID() string {
	if a == nil {
		return ""
	}
	return a.categoryID
}

func (a *Attempt) PolicyID() string {
	if a == nil {
		return ""
	}
	return a.policyID
}

func (a *Attempt) WindowID() string {
	if a == nil {
		return ""
	}
	return a.windowID
}

func (a *Attempt) ReservationID() string {
	if a == nil {
		return ""
	}
	return a.reservationID
}

func (a *Attempt) AttemptNo() int {
	if a == nil {
		return 0
	}
	return a.attemptNo
}

func (a *Attempt) State() AttemptState {
	if a == nil {
		return ""
	}
	return a.state
}

func (a *Attempt) Reason() string {
	if a == nil {
		return ""
	}
	return a.reason
}

func (a *Attempt) NextAttemptAt() (time.Time, bool) {
	if a == nil || a.nextAttemptAt == nil {
		return time.Time{}, false
	}
	return *a.nextAttemptAt, true
}

func (a *Attempt) MarkReady(now time.Time, reason string) error {
	if err := a.updateState(AttemptReady, reason); err != nil {
		return err
	}
	a.bumpVersion(now)
	return nil
}

func (a *Attempt) MarkRunning(now time.Time, reason string) error {
	if err := a.updateState(AttemptRunning, reason); err != nil {
		return err
	}
	a.bumpVersion(now)
	return nil
}

func (a *Attempt) MarkSucceeded(now time.Time, reason string) error {
	if err := a.updateState(AttemptSucceeded, reason); err != nil {
		return err
	}
	a.reason = reason
	a.nextAttemptAt = nil
	a.bumpVersion(now)
	return nil
}

func (a *Attempt) MarkFailed(now time.Time, reason string, nextAttemptAt *time.Time) error {
	if err := a.updateState(AttemptFailed, reason); err != nil {
		return err
	}
	a.nextAttemptAt = cloneTimePointer(nextAttemptAt)
	a.bumpVersion(now)
	return nil
}

func (a *Attempt) MarkBackoff(now time.Time, reason string, nextAttemptAt *time.Time) error {
	if err := a.updateState(AttemptBackoff, reason); err != nil {
		return err
	}
	a.nextAttemptAt = cloneTimePointer(nextAttemptAt)
	a.bumpVersion(now)
	return nil
}

func (a *Attempt) MarkTimedOut(now time.Time, reason string) error {
	if err := a.updateState(AttemptTimedOut, reason); err != nil {
		return err
	}
	a.bumpVersion(now)
	return nil
}

func (a *Attempt) MarkTerminated(now time.Time, reason string) error {
	if err := a.updateState(AttemptTerminated, reason); err != nil {
		return err
	}
	a.nextAttemptAt = nil
	a.bumpVersion(now)
	return nil
}

func (a *Attempt) updateState(to AttemptState, reason string) error {
	if a == nil {
		return ErrNilArgument
	}
	if err := ValidateAttemptTransition(a.state, to); err != nil {
		return err
	}
	a.state = to
	a.reason = reason
	return nil
}

func (a *Attempt) Clone() *Attempt {
	if a == nil {
		return nil
	}
	cp := *a
	cp.nextAttemptAt = cloneTimePointer(a.nextAttemptAt)
	return &cp
}
