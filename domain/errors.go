package domain

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidTransition                = errors.New("domain: invalid state transition")
	ErrBudgetExhausted                  = errors.New("domain: budget exhausted")
	ErrReservationExpired               = errors.New("domain: reservation expired")
	ErrLateResult                       = errors.New("domain: late result rejected")
	ErrVersionConflict                  = errors.New("domain: version conflict")
	ErrInvalidInvariant                 = errors.New("domain: invalid invariant")
	ErrReservationNotFound              = errors.New("domain: reservation not found")
	ErrAttemptNotReady                  = errors.New("domain: attempt not ready")
	ErrSuspensionRequired               = errors.New("domain: suspension required")
	ErrRecoveryCredentialExpired        = errors.New("domain: recovery credential expired")
	ErrInvalidBudgetPolicy              = errors.New("domain: invalid budget policy")
	ErrWindowClosed                     = errors.New("domain: window closed")
	ErrCannotReleaseConsumedReservation = errors.New("domain: cannot release consumed reservation")
	ErrNilArgument                      = errors.New("domain: nil argument")
)

type ErrorKind string

const (
	ErrorKindConflict        ErrorKind = "conflict"
	ErrorKindNotFound        ErrorKind = "not_found"
	ErrorKindInvalidArgument ErrorKind = "invalid_argument"
	ErrorKindPrecondition    ErrorKind = "precondition"
	ErrorKindExpired         ErrorKind = "expired"
	ErrorKindLateResult      ErrorKind = "late_result"
	ErrorKindBudget          ErrorKind = "budget"
)

type Error struct {
	Kind ErrorKind
	Op   string
	Err  error
}

func NewError(kind ErrorKind, op string, err error) error {
	return &Error{Kind: kind, Op: op, Err: err}
}

func (e *Error) Error() string {
	if e == nil {
		return "domain: <nil>"
	}
	switch {
	case e.Op != "" && e.Err != nil:
		return fmt.Sprintf("domain: %s: %v", e.Op, e.Err)
	case e.Op != "":
		return fmt.Sprintf("domain: %s", e.Op)
	case e.Err != nil:
		return fmt.Sprintf("domain: %v", e.Err)
	default:
		return "domain: unknown error"
	}
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *Error) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if t, ok := target.(*Error); ok {
		return e.Kind == t.Kind && e.Op == t.Op
	}
	return false
}
