package domain

import "time"

type Suspension struct {
	BaseAggregate
	categoryID     string
	attemptID      string
	reason         string
	suspendedUntil time.Time
	liftedAt       *time.Time
}

func NewSuspension(id, categoryID, attemptID, reason string, suspendedUntil time.Time, now time.Time) (*Suspension, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	if categoryID == "" || attemptID == "" {
		return nil, NewError(ErrorKindInvalidArgument, "NewSuspension", ErrInvalidInvariant)
	}
	if reason == "" {
		return nil, NewError(ErrorKindInvalidArgument, "NewSuspension", ErrInvalidInvariant)
	}
	if suspendedUntil.Before(now) {
		return nil, NewError(ErrorKindInvalidArgument, "NewSuspension", ErrInvalidInvariant)
	}
	s := &Suspension{
		categoryID:     categoryID,
		attemptID:      attemptID,
		reason:         reason,
		suspendedUntil: suspendedUntil,
	}
	s.setID(id)
	s.setVersion(1)
	s.setCreatedAt(now)
	s.setUpdatedAt(now)
	return s, nil
}

func (s *Suspension) CategoryID() string {
	if s == nil {
		return ""
	}
	return s.categoryID
}

func (s *Suspension) AttemptID() string {
	if s == nil {
		return ""
	}
	return s.attemptID
}

func (s *Suspension) Reason() string {
	if s == nil {
		return ""
	}
	return s.reason
}

func (s *Suspension) SuspendedUntil() time.Time {
	if s == nil {
		return time.Time{}
	}
	return s.suspendedUntil
}

func (s *Suspension) IsActive(now time.Time) bool {
	if s == nil {
		return false
	}
	return s.liftedAt == nil && now.Before(s.suspendedUntil)
}

func (s *Suspension) Lift(now time.Time) error {
	if s == nil {
		return ErrNilArgument
	}
	t := now
	s.liftedAt = &t
	s.bumpVersion(now)
	return nil
}

func (s *Suspension) Clone() *Suspension {
	if s == nil {
		return nil
	}
	cp := *s
	cp.liftedAt = cloneTimePointer(s.liftedAt)
	return &cp
}
