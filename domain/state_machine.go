package domain

import "fmt"

type AttemptState string

const (
	AttemptReserved   AttemptState = "RESERVED"
	AttemptReady      AttemptState = "READY"
	AttemptRunning    AttemptState = "RUNNING"
	AttemptSucceeded  AttemptState = "SUCCEEDED"
	AttemptFailed     AttemptState = "FAILED"
	AttemptBackoff    AttemptState = "BACKOFF"
	AttemptTimedOut   AttemptState = "TIMED_OUT"
	AttemptTerminated AttemptState = "TERMINATED"
)

var attemptTransitions = map[AttemptState]map[AttemptState]bool{
	AttemptReserved:   {AttemptReady: true, AttemptTimedOut: true, AttemptTerminated: true},
	AttemptReady:      {AttemptRunning: true, AttemptTimedOut: true, AttemptTerminated: true},
	AttemptRunning:    {AttemptSucceeded: true, AttemptFailed: true, AttemptTimedOut: true, AttemptTerminated: true},
	AttemptFailed:     {AttemptBackoff: true, AttemptTerminated: true},
	AttemptBackoff:    {AttemptReady: true, AttemptTerminated: true},
	AttemptTimedOut:   {AttemptTerminated: true},
	AttemptSucceeded:  {},
	AttemptTerminated: {},
}

func ValidateAttemptTransition(from, to AttemptState) error {
	if to == "" {
		return NewError(ErrorKindInvalidArgument, "ValidateAttemptTransition", ErrInvalidTransition)
	}
	allowed, ok := attemptTransitions[from]
	if !ok {
		return NewError(ErrorKindInvalidArgument, "ValidateAttemptTransition", fmt.Errorf("%w: unknown source state %q", ErrInvalidTransition, from))
	}
	if allowed[to] {
		return nil
	}
	return NewError(ErrorKindInvalidArgument, "ValidateAttemptTransition", fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to))
}

type ReservationState string

const (
	ReservationActive   ReservationState = "ACTIVE"
	ReservationConsumed ReservationState = "CONSUMED"
	ReservationReleased ReservationState = "RELEASED"
	ReservationExpired  ReservationState = "EXPIRED"
)

var reservationTransitions = map[ReservationState]map[ReservationState]bool{
	ReservationActive:   {ReservationConsumed: true, ReservationReleased: true, ReservationExpired: true},
	ReservationConsumed: {},
	ReservationReleased: {},
	ReservationExpired:  {},
}

func ValidateReservationTransition(from, to ReservationState) error {
	if to == "" {
		return NewError(ErrorKindInvalidArgument, "ValidateReservationTransition", ErrInvalidTransition)
	}
	allowed, ok := reservationTransitions[from]
	if !ok {
		return NewError(ErrorKindInvalidArgument, "ValidateReservationTransition", fmt.Errorf("%w: unknown source state %q", ErrInvalidTransition, from))
	}
	if allowed[to] {
		return nil
	}
	return NewError(ErrorKindInvalidArgument, "ValidateReservationTransition", fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to))
}
