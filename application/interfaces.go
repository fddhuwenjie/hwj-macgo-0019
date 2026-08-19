package application

import (
	"context"
	"time"
)

type Clock interface {
	Now() time.Time
}

type AuditSink interface {
	Record(ctx context.Context, event AuditEvent) error
	Export(ctx context.Context) ([]AuditEvent, error)
}

type AuditEvent struct {
	Seq      int64
	At       time.Time
	Actor    string
	Action   string
	Entity   string
	EntityID string
	Detail   string
	Checksum string
}

type Store interface {
	Begin(ctx context.Context) (Tx, error)
	InTx(ctx context.Context, fn func(tx Tx) error) error
}

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error

	Categories() CategoryRepository
	Attempts() AttemptRepository
	Windows() WindowRepository
	Reservations() ReservationRepository
	Results() ResultRepository
}

type CategoryRepository interface {
	FindByID(ctx context.Context, id string) (Category, error)
	Save(ctx context.Context, c Category) error
	List(ctx context.Context) ([]Category, error)
}

type AttemptRepository interface {
	FindByID(ctx context.Context, id string) (Attempt, error)
	FindByIdempotencyKey(ctx context.Context, key string) (Attempt, error)
	Save(ctx context.Context, a Attempt) error
	ListByCategory(ctx context.Context, categoryID string, after string, limit int) ([]Attempt, error)
	ListRetryable(ctx context.Context, categoryID string, before time.Time, limit int) ([]Attempt, error)
}

type WindowRepository interface {
	FindByCategory(ctx context.Context, categoryID string) (Window, error)
	Save(ctx context.Context, w Window) error
}

type ReservationRepository interface {
	FindByID(ctx context.Context, id string) (Reservation, error)
	Save(ctx context.Context, r Reservation) error
	ListByCategory(ctx context.Context, categoryID string) ([]Reservation, error)
}

type ResultRepository interface {
	Save(ctx context.Context, r Result) error
	FindByAttempt(ctx context.Context, attemptID string) (Result, error)
}

type Category interface {
	ID() string
	Version() int64
	State() string
	BudgetPolicy() BudgetPolicy
	CreatedAt() time.Time
	UpdatedAt() time.Time
}

type Attempt interface {
	ID() string
	CategoryID() string
	ReservationID() string
	State() string
	Version() int64
	FailureCount() int
	NextAttemptAt() time.Time
	Reason() string
	CreatedAt() time.Time
	UpdatedAt() time.Time
}

type Window interface {
	CategoryID() string
	Version() int64
	Used() int
	Reserved() int
	StartAt() time.Time
	EndAt() time.Time
	CreatedAt() time.Time
	UpdatedAt() time.Time
}

type Reservation interface {
	ID() string
	CategoryID() string
	State() string
	Version() int64
	ExpiresAt() time.Time
	CreatedAt() time.Time
	UpdatedAt() time.Time
}

type Result interface {
	AttemptID() string
	Success() bool
	Terminated() bool
	RecordedAt() time.Time
	Version() int64
}
