package domain

import "time"

type BackoffConfig struct {
	Initial     time.Duration
	Max         time.Duration
	Multiplier  float64
	JitterRatio float64
}

func (c BackoffConfig) Validate() error {
	if c.Initial <= 0 {
		return NewError(ErrorKindInvalidArgument, "BackoffConfig.Validate", ErrInvalidBudgetPolicy)
	}
	if c.Max <= 0 {
		return NewError(ErrorKindInvalidArgument, "BackoffConfig.Validate", ErrInvalidBudgetPolicy)
	}
	if c.Multiplier <= 0 {
		return NewError(ErrorKindInvalidArgument, "BackoffConfig.Validate", ErrInvalidBudgetPolicy)
	}
	if c.JitterRatio < 0 || c.JitterRatio > 1 {
		return NewError(ErrorKindInvalidArgument, "BackoffConfig.Validate", ErrInvalidBudgetPolicy)
	}
	return nil
}

func (c BackoffConfig) Delay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	d := c.Initial
	for i := 1; i < attempt; i++ {
		d = time.Duration(float64(d) * c.Multiplier)
		if d <= 0 {
			return c.Max
		}
		if d >= c.Max {
			return c.Max
		}
	}
	if c.JitterRatio > 0 && d > 0 {
		jitter := time.Duration(float64(d) * c.JitterRatio)
		d += jitter
		if d > c.Max {
			d = c.Max
		}
	}
	return d
}

type BackoffPlan struct {
	BaseAggregate
	attemptID     string
	categoryID    string
	config        BackoffConfig
	attemptNo     int
	nextAttemptAt time.Time
	lastDelay     time.Duration
}

func NewBackoffPlan(id, attemptID, categoryID string, config BackoffConfig, attemptNo int, now time.Time) (*BackoffPlan, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	if attemptID == "" || categoryID == "" {
		return nil, NewError(ErrorKindInvalidArgument, "NewBackoffPlan", ErrInvalidInvariant)
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if attemptNo < 0 {
		return nil, NewError(ErrorKindInvalidArgument, "NewBackoffPlan", ErrInvalidInvariant)
	}
	p := &BackoffPlan{
		attemptID:  attemptID,
		categoryID: categoryID,
		config:     config,
		attemptNo:  attemptNo,
		lastDelay:  config.Delay(attemptNo),
	}
	p.nextAttemptAt = now.Add(p.lastDelay)
	p.setID(id)
	p.setVersion(1)
	p.setCreatedAt(now)
	p.setUpdatedAt(now)
	return p, nil
}

func (p *BackoffPlan) AttemptID() string {
	if p == nil {
		return ""
	}
	return p.attemptID
}

func (p *BackoffPlan) CategoryID() string {
	if p == nil {
		return ""
	}
	return p.categoryID
}

func (p *BackoffPlan) Config() BackoffConfig {
	if p == nil {
		return BackoffConfig{}
	}
	return p.config
}

func (p *BackoffPlan) AttemptNo() int {
	if p == nil {
		return 0
	}
	return p.attemptNo
}

func (p *BackoffPlan) NextAttemptAt() time.Time {
	if p == nil {
		return time.Time{}
	}
	return p.nextAttemptAt
}

func (p *BackoffPlan) LastDelay() time.Duration {
	if p == nil {
		return 0
	}
	return p.lastDelay
}

func (p *BackoffPlan) Advance(now time.Time) (time.Time, error) {
	if p == nil {
		return time.Time{}, ErrNilArgument
	}
	if now.Before(p.nextAttemptAt) {
		return time.Time{}, NewError(ErrorKindPrecondition, "BackoffPlan.Advance", ErrAttemptNotReady)
	}
	p.attemptNo++
	p.lastDelay = p.config.Delay(p.attemptNo)
	p.nextAttemptAt = now.Add(p.lastDelay)
	p.bumpVersion(now)
	return p.nextAttemptAt, nil
}

func (p *BackoffPlan) Clone() *BackoffPlan {
	if p == nil {
		return nil
	}
	cp := *p
	return &cp
}
