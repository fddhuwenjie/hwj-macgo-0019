package application

import "time"

type appWindowRecord struct {
	categoryID string
	version    int64
	used       int
	reserved   int
	startAt    time.Time
	endAt      time.Time
	createdAt  time.Time
	updatedAt  time.Time
}

func newAppWindowRecord(w Window) *appWindowRecord {
	return &appWindowRecord{
		categoryID: w.CategoryID(),
		version:    w.Version(),
		used:       w.Used(),
		reserved:   w.Reserved(),
		startAt:    w.StartAt(),
		endAt:      w.EndAt(),
		createdAt:  w.CreatedAt(),
		updatedAt:  w.UpdatedAt(),
	}
}

func (r *appWindowRecord) CategoryID() string   { return r.categoryID }
func (r *appWindowRecord) Version() int64       { return r.version }
func (r *appWindowRecord) Used() int            { return r.used }
func (r *appWindowRecord) Reserved() int        { return r.reserved }
func (r *appWindowRecord) StartAt() time.Time   { return r.startAt }
func (r *appWindowRecord) EndAt() time.Time     { return r.endAt }
func (r *appWindowRecord) CreatedAt() time.Time { return r.createdAt }
func (r *appWindowRecord) UpdatedAt() time.Time { return r.updatedAt }

type appAttemptRecord struct {
	id            string
	categoryID    string
	reservationID string
	state         string
	version       int64
	failureCount  int
	nextAttemptAt time.Time
	reason        string
	createdAt     time.Time
	updatedAt     time.Time
}

func newAppAttemptRecord(a Attempt) *appAttemptRecord {
	return &appAttemptRecord{
		id:            a.ID(),
		categoryID:    a.CategoryID(),
		reservationID: a.ReservationID(),
		state:         a.State(),
		version:       a.Version(),
		failureCount:  a.FailureCount(),
		nextAttemptAt: a.NextAttemptAt(),
		reason:        a.Reason(),
		createdAt:     a.CreatedAt(),
		updatedAt:     a.UpdatedAt(),
	}
}

func (r *appAttemptRecord) ID() string               { return r.id }
func (r *appAttemptRecord) CategoryID() string       { return r.categoryID }
func (r *appAttemptRecord) ReservationID() string    { return r.reservationID }
func (r *appAttemptRecord) State() string            { return r.state }
func (r *appAttemptRecord) Version() int64           { return r.version }
func (r *appAttemptRecord) FailureCount() int        { return r.failureCount }
func (r *appAttemptRecord) NextAttemptAt() time.Time { return r.nextAttemptAt }
func (r *appAttemptRecord) Reason() string           { return r.reason }
func (r *appAttemptRecord) CreatedAt() time.Time     { return r.createdAt }
func (r *appAttemptRecord) UpdatedAt() time.Time     { return r.updatedAt }

type appReservationRecord struct {
	id         string
	categoryID string
	state      string
	version    int64
	expiresAt  time.Time
	createdAt  time.Time
	updatedAt  time.Time
}

func newAppReservationRecord(r Reservation) *appReservationRecord {
	return &appReservationRecord{
		id:         r.ID(),
		categoryID: r.CategoryID(),
		state:      r.State(),
		version:    r.Version(),
		expiresAt:  r.ExpiresAt(),
		createdAt:  r.CreatedAt(),
		updatedAt:  r.UpdatedAt(),
	}
}

func (r *appReservationRecord) ID() string           { return r.id }
func (r *appReservationRecord) CategoryID() string   { return r.categoryID }
func (r *appReservationRecord) State() string        { return r.state }
func (r *appReservationRecord) Version() int64       { return r.version }
func (r *appReservationRecord) ExpiresAt() time.Time { return r.expiresAt }
func (r *appReservationRecord) CreatedAt() time.Time { return r.createdAt }
func (r *appReservationRecord) UpdatedAt() time.Time { return r.updatedAt }

type appResultRecord struct {
	attemptID  string
	success    bool
	terminated bool
	recordedAt time.Time
	version    int64
}

func newAppResultRecord(r Result) *appResultRecord {
	return &appResultRecord{
		attemptID:  r.AttemptID(),
		success:    r.Success(),
		terminated: r.Terminated(),
		recordedAt: r.RecordedAt(),
		version:    r.Version(),
	}
}

func (r *appResultRecord) AttemptID() string     { return r.attemptID }
func (r *appResultRecord) Success() bool         { return r.success }
func (r *appResultRecord) Terminated() bool      { return r.terminated }
func (r *appResultRecord) RecordedAt() time.Time { return r.recordedAt }
func (r *appResultRecord) Version() int64        { return r.version }
