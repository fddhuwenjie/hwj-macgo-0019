package domain

import "time"

type Result struct {
	BaseAggregate
	attemptID     string
	success       bool
	failureReason string
	outcome       string
	observedAt    time.Time
	metadata      map[string]string
}

func NewResult(id, attemptID, outcome string, success bool, failureReason string, observedAt time.Time, metadata map[string]string, now time.Time) (*Result, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	if attemptID == "" {
		return nil, NewError(ErrorKindInvalidArgument, "NewResult", ErrInvalidInvariant)
	}
	if observedAt.IsZero() {
		observedAt = now
	}
	r := &Result{
		attemptID:     attemptID,
		success:       success,
		failureReason: failureReason,
		outcome:       outcome,
		observedAt:    observedAt,
		metadata:      cloneStringMap(metadata),
	}
	r.setID(id)
	r.setVersion(1)
	r.setCreatedAt(now)
	r.setUpdatedAt(now)
	return r, nil
}

func (r *Result) AttemptID() string {
	if r == nil {
		return ""
	}
	return r.attemptID
}

func (r *Result) Success() bool {
	if r == nil {
		return false
	}
	return r.success
}

func (r *Result) FailureReason() string {
	if r == nil {
		return ""
	}
	return r.failureReason
}

func (r *Result) Outcome() string {
	if r == nil {
		return ""
	}
	return r.outcome
}

func (r *Result) ObservedAt() time.Time {
	if r == nil {
		return time.Time{}
	}
	return r.observedAt
}

func (r *Result) Metadata() map[string]string {
	if r == nil {
		return nil
	}
	return cloneStringMap(r.metadata)
}

func (r *Result) LateFor(deadline time.Time) bool {
	if r == nil {
		return true
	}
	return r.observedAt.After(deadline)
}

func (r *Result) Clone() *Result {
	if r == nil {
		return nil
	}
	cp := *r
	cp.metadata = cloneStringMap(r.metadata)
	return &cp
}
