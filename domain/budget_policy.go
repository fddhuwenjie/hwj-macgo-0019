package domain

import "time"

type BudgetPolicy struct {
	BaseAggregate
	categoryID      string
	maxAttempts     int
	maxReservations int
	windowDuration  time.Duration
	backoff         BackoffConfig
}

func NewBudgetPolicy(id, categoryID string, maxAttempts int, window time.Duration, backoff BackoffConfig, now time.Time) (*BudgetPolicy, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	if categoryID == "" {
		return nil, NewError(ErrorKindInvalidArgument, "NewBudgetPolicy", ErrInvalidInvariant)
	}
	if maxAttempts <= 0 {
		return nil, NewError(ErrorKindInvalidArgument, "NewBudgetPolicy", ErrInvalidBudgetPolicy)
	}
	if window <= 0 {
		return nil, NewError(ErrorKindInvalidArgument, "NewBudgetPolicy", ErrInvalidBudgetPolicy)
	}
	if err := backoff.Validate(); err != nil {
		return nil, err
	}
	p := &BudgetPolicy{
		categoryID:      categoryID,
		maxAttempts:     maxAttempts,
		maxReservations: maxAttempts,
		windowDuration:  window,
		backoff:         backoff,
	}
	p.setID(id)
	p.setVersion(1)
	p.setCreatedAt(now)
	p.setUpdatedAt(now)
	return p, nil
}

func (p *BudgetPolicy) CategoryID() string {
	if p == nil {
		return ""
	}
	return p.categoryID
}

func (p *BudgetPolicy) MaxAttempts() int {
	if p == nil {
		return 0
	}
	return p.maxAttempts
}

func (p *BudgetPolicy) MaxReservations() int {
	if p == nil {
		return 0
	}
	return p.maxReservations
}

func (p *BudgetPolicy) WindowDuration() time.Duration {
	if p == nil {
		return 0
	}
	return p.windowDuration
}

func (p *BudgetPolicy) Backoff() BackoffConfig {
	if p == nil {
		return BackoffConfig{}
	}
	return p.backoff
}

func (p *BudgetPolicy) Resize(maxAttempts int, window time.Duration, backoff BackoffConfig, now time.Time) error {
	if p == nil {
		return ErrNilArgument
	}
	if maxAttempts <= 0 {
		return NewError(ErrorKindInvalidArgument, "BudgetPolicy.Resize", ErrInvalidBudgetPolicy)
	}
	if window <= 0 {
		return NewError(ErrorKindInvalidArgument, "BudgetPolicy.Resize", ErrInvalidBudgetPolicy)
	}
	if err := backoff.Validate(); err != nil {
		return err
	}
	p.maxAttempts = maxAttempts
	p.maxReservations = maxAttempts
	p.windowDuration = window
	p.backoff = backoff
	p.bumpVersion(now)
	return nil
}

func (p *BudgetPolicy) Clone() *BudgetPolicy {
	if p == nil {
		return nil
	}
	cp := *p
	return &cp
}
