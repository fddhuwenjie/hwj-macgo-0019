package scheduler

import (
	"errors"
	"math"
	"time"
)

// BackoffPolicy drives exponential retry delays.
type BackoffPolicy struct {
	Initial time.Duration
	Max     time.Duration
	Factor  float64
}

func (p BackoffPolicy) NextDelay(attempt int) time.Duration {
	if p.Initial <= 0 {
		p.Initial = time.Second
	}
	if p.Max <= 0 {
		p.Max = time.Minute
	}
	if p.Factor <= 1 {
		p.Factor = 2.0
	}
	if attempt < 1 {
		attempt = 1
	}
	d := float64(p.Initial) * math.Pow(p.Factor, float64(attempt-1))
	if d > float64(p.Max) {
		d = float64(p.Max)
	}
	return time.Duration(d)
}

// TerminationError marks an error as permanently terminating a task.
type TerminationError struct {
	Err       error
	Permanent bool
}

func (e TerminationError) Error() string {
	return e.Err.Error()
}

func (e TerminationError) Unwrap() error {
	return e.Err
}

func IsTerminal(err error) bool {
	var te TerminationError
	if errors.As(err, &te) {
		return te.Permanent
	}
	return false
}
