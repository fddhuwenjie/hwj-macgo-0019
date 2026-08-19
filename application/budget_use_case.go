package application

import (
	"context"
	"errors"
	"fmt"
)

var (
	budgetUseCaseErrExceeded         = errors.New("application: budget exceeded")
	budgetUseCaseErrReservedExceeded = errors.New("application: reservation budget exceeded")
)

type BudgetUseCase struct {
	engine *Engine
}

func NewBudgetUseCase(engine *Engine) *BudgetUseCase {
	return &BudgetUseCase{engine: engine}
}

func (uc *BudgetUseCase) RemainingBudget(ctx context.Context, categoryID string) (BudgetView, error) {
	var view BudgetView
	err := withTx(ctx, uc.engine.Store(), func(tx Tx) error {
		cat, err := tx.Categories().FindByID(ctx, categoryID)
		if err != nil {
			return err
		}
		win, err := tx.Windows().FindByCategory(ctx, categoryID)
		if err != nil {
			return err
		}
		policy := cat.BudgetPolicy()
		used := win.Used()
		reserved := win.Reserved()
		remaining := policy.MaxAttempts - used - reserved
		if remaining < 0 {
			remaining = 0
		}
		view = BudgetView{
			CategoryID:  categoryID,
			Used:        used,
			Reserved:    reserved,
			Remaining:   remaining,
			WindowStart: win.StartAt(),
			WindowEnd:   win.EndAt(),
		}
		return nil
	})
	return view, err
}

func (uc *BudgetUseCase) ReserveCapacity(ctx context.Context, categoryID string, count int) error {
	if count <= 0 {
		return fmt.Errorf("application: reserve count must be positive")
	}
	return withTx(ctx, uc.engine.Store(), func(tx Tx) error {
		return reserveCapacityInTx(ctx, tx, uc.engine.Clock(), categoryID, count)
	})
}

func (uc *BudgetUseCase) ReleaseReservation(ctx context.Context, categoryID, reservationID string) error {
	return withTx(ctx, uc.engine.Store(), func(tx Tx) error {
		return releaseReservationInTx(ctx, tx, uc.engine.Clock(), categoryID, reservationID)
	})
}

func reserveCapacityInTx(ctx context.Context, tx Tx, clock Clock, categoryID string, count int) error {
	if count <= 0 {
		return fmt.Errorf("application: reserve count must be positive")
	}
	cat, err := tx.Categories().FindByID(ctx, categoryID)
	if err != nil {
		return err
	}
	win, err := tx.Windows().FindByCategory(ctx, categoryID)
	if err != nil {
		return err
	}
	policy := cat.BudgetPolicy()
	now := clock.Now()
	if now.Before(win.StartAt()) || now.After(win.EndAt()) {
		return fmt.Errorf("application: window is not active")
	}
	if win.Used()+win.Reserved()+count > policy.MaxAttempts {
		return budgetUseCaseErrExceeded
	}
	if win.Reserved()+count > policy.MaxReservations {
		return budgetUseCaseErrReservedExceeded
	}
	rec := newAppWindowRecord(win)
	rec.reserved += count
	rec.version++
	rec.updatedAt = now
	return tx.Windows().Save(ctx, rec)
}

func releaseReservationInTx(ctx context.Context, tx Tx, clock Clock, categoryID, reservationID string) error {
	res, err := tx.Reservations().FindByID(ctx, reservationID)
	if err != nil {
		return err
	}
	if res.CategoryID() != categoryID {
		return fmt.Errorf("application: reservation does not belong to category")
	}
	if res.State() == "released" || res.State() == "expired" {
		return nil
	}
	win, err := tx.Windows().FindByCategory(ctx, categoryID)
	if err != nil {
		return err
	}
	now := clock.Now()
	rec := newAppWindowRecord(win)
	if now.Before(win.EndAt()) && now.After(win.StartAt()) && win.Reserved() > 0 {
		rec.reserved--
	}
	rec.updatedAt = now
	rec.version++
	if err := tx.Windows().Save(ctx, rec); err != nil {
		return err
	}
	rr := newAppReservationRecord(res)
	rr.state = "released"
	rr.version++
	rr.updatedAt = now
	return tx.Reservations().Save(ctx, rr)
}
