package domain

import "time"

func ValidateBudgetWindow(policy *BudgetPolicy, window *Window, now time.Time) error {
	if policy == nil || window == nil {
		return ErrNilArgument
	}
	if window.PolicyID() != policy.ID() {
		return NewError(ErrorKindPrecondition, "ValidateBudgetWindow", ErrInvalidInvariant)
	}
	if window.IsExpired(now) {
		return NewError(ErrorKindExpired, "ValidateBudgetWindow", ErrWindowClosed)
	}
	if window.Limit() != policy.MaxAttempts() {
		return NewError(ErrorKindPrecondition, "ValidateBudgetWindow", ErrInvalidInvariant)
	}
	if window.Used()+window.Reserved() > policy.MaxAttempts() {
		return NewError(ErrorKindBudget, "ValidateBudgetWindow", ErrBudgetExhausted)
	}
	return nil
}

func ValidateReservationOnWindow(window *Window, reservation *Reservation, now time.Time) error {
	if window == nil || reservation == nil {
		return ErrNilArgument
	}
	if reservation.WindowID() != window.ID() {
		return NewError(ErrorKindPrecondition, "ValidateReservationOnWindow", ErrInvalidInvariant)
	}
	if reservation.Amount() <= 0 {
		return NewError(ErrorKindPrecondition, "ValidateReservationOnWindow", ErrInvalidInvariant)
	}
	if reservation.State() == ReservationActive && reservation.IsExpired(now) {
		return NewError(ErrorKindExpired, "ValidateReservationOnWindow", ErrReservationExpired)
	}
	if window.Remaining() < reservation.Amount() && reservation.State() == ReservationActive {
		return NewError(ErrorKindBudget, "ValidateReservationOnWindow", ErrBudgetExhausted)
	}
	return nil
}

func EnsureSuccessReleasesReservation(attempt *Attempt, reservation *Reservation) error {
	if attempt == nil || reservation == nil {
		return ErrNilArgument
	}
	if attempt.State() != AttemptSucceeded {
		return NewError(ErrorKindPrecondition, "EnsureSuccessReleasesReservation", ErrAttemptNotReady)
	}
	if reservation.State() != ReservationReleased {
		return NewError(ErrorKindPrecondition, "EnsureSuccessReleasesReservation", ErrInvalidInvariant)
	}
	return nil
}

func EnsureLateResultRejected(result *Result, window *Window, now time.Time) error {
	if result == nil || window == nil {
		return ErrNilArgument
	}
	if result.ObservedAt().After(window.ClosesAt()) {
		return NewError(ErrorKindLateResult, "EnsureLateResultRejected", ErrLateResult)
	}
	return nil
}

func ValidateRecoveryCredential(credential *RecoveryCredential, categoryID string, now time.Time) error {
	if credential == nil {
		return ErrNilArgument
	}
	if credential.CategoryID() != categoryID {
		return NewError(ErrorKindPrecondition, "ValidateRecoveryCredential", ErrInvalidInvariant)
	}
	if credential.IsExpired(now) {
		return NewError(ErrorKindExpired, "ValidateRecoveryCredential", ErrRecoveryCredentialExpired)
	}
	if credential.IsRedeemed() {
		return NewError(ErrorKindPrecondition, "ValidateRecoveryCredential", ErrInvalidTransition)
	}
	return nil
}

func ValidateBackoffLimit(streak *FailureStreak, maxConsecutiveFailures int) error {
	if streak == nil {
		return ErrNilArgument
	}
	if maxConsecutiveFailures <= 0 {
		return NewError(ErrorKindInvalidArgument, "ValidateBackoffLimit", ErrInvalidBudgetPolicy)
	}
	if streak.Count() >= maxConsecutiveFailures {
		return NewError(ErrorKindPrecondition, "ValidateBackoffLimit", ErrSuspensionRequired)
	}
	return nil
}
