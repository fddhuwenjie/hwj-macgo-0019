package application

import (
	"context"
	"errors"
	"fmt"
)

var (
	attemptUseCaseErrDuplicate          = errors.New("application: duplicate idempotency key")
	attemptUseCaseErrReservationExpired = errors.New("application: reservation expired")
)

type AttemptUseCase struct {
	engine *Engine
}

func NewAttemptUseCase(engine *Engine) *AttemptUseCase {
	return &AttemptUseCase{engine: engine}
}

func (uc *AttemptUseCase) Apply(ctx context.Context, cmd AttemptCommand) (Attempt, error) {
	var saved Attempt
	err := withTx(ctx, uc.engine.Store(), func(tx Tx) error {
		if cmd.IdempotencyKey != "" {
			existing, err := tx.Attempts().FindByIdempotencyKey(ctx, cmd.IdempotencyKey)
			if err == nil && existing != nil {
				saved = existing
				return attemptUseCaseErrDuplicate
			}
		}
		res, err := tx.Reservations().FindByID(ctx, cmd.ReservationID)
		if err != nil {
			return err
		}
		if res.CategoryID() != cmd.CategoryID {
			return fmt.Errorf("application: reservation category mismatch")
		}
		now := uc.engine.Clock().Now()
		if now.After(res.ExpiresAt()) {
			return attemptUseCaseErrReservationExpired
		}
		if err := reserveCapacityInTx(ctx, tx, uc.engine.Clock(), cmd.CategoryID, 1); err != nil {
			return err
		}
		id := cmd.IdempotencyKey
		if id == "" {
			id = fmt.Sprintf("app-%d", now.UnixNano())
		}
		rec := &appAttemptRecord{
			id:            id,
			categoryID:    cmd.CategoryID,
			reservationID: cmd.ReservationID,
			state:         "reserved",
			version:       1,
			failureCount:  0,
			nextAttemptAt: now,
			createdAt:     now,
			updatedAt:     now,
		}
		if err := tx.Attempts().Save(ctx, rec); err != nil {
			return err
		}
		saved = rec
		uc.engine.record(ctx, "attempt.apply", "attempt", id, "reserved attempt")
		return nil
	})
	if errors.Is(err, attemptUseCaseErrDuplicate) && saved != nil {
		return saved, nil
	}
	return saved, err
}

func (uc *AttemptUseCase) RecordSuccess(ctx context.Context, attemptID string) error {
	return withTx(ctx, uc.engine.Store(), func(tx Tx) error {
		a, err := tx.Attempts().FindByID(ctx, attemptID)
		if err != nil {
			return err
		}
		if a.State() == "success" || a.State() == "terminated" {
			return nil
		}
		now := uc.engine.Clock().Now()
		ar := newAppAttemptRecord(a)
		ar.state = "success"
		ar.version++
		ar.updatedAt = now
		ar.reason = ""
		if err := tx.Attempts().Save(ctx, ar); err != nil {
			return err
		}
		if err := tx.Results().Save(ctx, &appResultRecord{
			attemptID:  attemptID,
			success:    true,
			terminated: false,
			recordedAt: now,
			version:    1,
		}); err != nil {
			return err
		}
		if ar.reservationID != "" {
			if err := releaseReservationInTx(ctx, tx, uc.engine.Clock(), ar.categoryID, ar.reservationID); err != nil {
				return err
			}
		}
		uc.engine.record(ctx, "attempt.success", "attempt", attemptID, "success")
		return nil
	})
}

func (uc *AttemptUseCase) RecordFailure(ctx context.Context, attemptID string, reason string) error {
	return withTx(ctx, uc.engine.Store(), func(tx Tx) error {
		a, err := tx.Attempts().FindByID(ctx, attemptID)
		if err != nil {
			return err
		}
		if a.State() == "success" || a.State() == "terminated" {
			return fmt.Errorf("application: outcome already recorded")
		}
		ar := newAppAttemptRecord(a)
		ar.state = "failed"
		ar.failureCount++
		ar.reason = reason
		ar.version++
		ar.updatedAt = uc.engine.Clock().Now()
		if err := tx.Attempts().Save(ctx, ar); err != nil {
			return err
		}
		if err := tx.Results().Save(ctx, &appResultRecord{
			attemptID:  attemptID,
			success:    false,
			terminated: false,
			recordedAt: uc.engine.Clock().Now(),
			version:    1,
		}); err != nil {
			return err
		}
		uc.engine.record(ctx, "attempt.failure", "attempt", attemptID, reason)
		return nil
	})
}
