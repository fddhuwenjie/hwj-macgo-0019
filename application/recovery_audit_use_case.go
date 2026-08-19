package application

import (
	"context"
	"fmt"
)

type RecoveryUseCase struct {
	engine *Engine
}

func NewRecoveryUseCase(engine *Engine) *RecoveryUseCase {
	return &RecoveryUseCase{engine: engine}
}

func (uc *RecoveryUseCase) Reconcile(ctx context.Context) error {
	return withTx(ctx, uc.engine.Store(), func(tx Tx) error {
		categories, err := tx.Categories().List(ctx)
		if err != nil {
			return err
		}
		for _, cat := range categories {
			win, err := tx.Windows().FindByCategory(ctx, cat.ID())
			if err != nil {
				return err
			}
			policy := cat.BudgetPolicy()
			if win.Used() < 0 || win.Reserved() < 0 || win.Used()+win.Reserved() > policy.MaxAttempts {
				return fmt.Errorf("application: budget invariant violation for category %s", cat.ID())
			}
			reservations, err := tx.Reservations().ListByCategory(ctx, cat.ID())
			if err != nil {
				return err
			}
			active := 0
			for _, r := range reservations {
				if r.State() != "reserved" {
					continue
				}
				if r.ExpiresAt().After(win.StartAt()) && r.ExpiresAt().Before(win.EndAt()) {
					active++
				}
			}
			if active > win.Reserved() {
				return fmt.Errorf("application: reserved count mismatch for category %s", cat.ID())
			}
		}
		return nil
	})
}

func (uc *RecoveryUseCase) ExportAudit(ctx context.Context) ([]AuditEvent, error) {
	if uc.engine.Audit() == nil {
		return nil, fmt.Errorf("application: audit sink is not configured")
	}
	return uc.engine.Audit().Export(ctx)
}
