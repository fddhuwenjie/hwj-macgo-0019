package application

import "context"

type WindowUseCase struct {
	engine *Engine
}

func NewWindowUseCase(engine *Engine) *WindowUseCase {
	return &WindowUseCase{engine: engine}
}

func (uc *WindowUseCase) RollWindow(ctx context.Context, categoryID string) (WindowView, error) {
	var view WindowView
	err := withTx(ctx, uc.engine.Store(), func(tx Tx) error {
		cat, err := tx.Categories().FindByID(ctx, categoryID)
		if err != nil {
			return err
		}
		win, err := tx.Windows().FindByCategory(ctx, categoryID)
		if err != nil {
			return err
		}
		now := uc.engine.Clock().Now()
		if now.Before(win.EndAt()) {
			view = WindowView{
				CategoryID: categoryID,
				Used:       win.Used(),
				Reserved:   win.Reserved(),
				StartAt:    win.StartAt(),
				EndAt:      win.EndAt(),
			}
			return nil
		}
		policy := cat.BudgetPolicy()
		rec := newAppWindowRecord(win)
		rec.startAt = win.EndAt()
		rec.endAt = win.EndAt().Add(policy.WindowDuration)
		rec.used = 0
		rec.reserved = 0
		rec.version++
		rec.updatedAt = now
		if err := tx.Windows().Save(ctx, rec); err != nil {
			return err
		}
		reservations, err := tx.Reservations().ListByCategory(ctx, categoryID)
		if err != nil {
			return err
		}
		for _, r := range reservations {
			if r.State() == "reserved" && r.ExpiresAt().Before(rec.startAt) {
				rr := newAppReservationRecord(r)
				rr.state = "expired"
				rr.version++
				rr.updatedAt = now
				if err := tx.Reservations().Save(ctx, rr); err != nil {
					return err
				}
			}
		}
		view = WindowView{
			CategoryID: categoryID,
			Used:       rec.used,
			Reserved:   rec.reserved,
			StartAt:    rec.startAt,
			EndAt:      rec.endAt,
		}
		uc.engine.record(ctx, "window.roll", "window", categoryID, "window rolled")
		return nil
	})
	return view, err
}

func (uc *WindowUseCase) HandleLateResult(ctx context.Context, attemptID string, success bool) error {
	return withTx(ctx, uc.engine.Store(), func(tx Tx) error {
		a, err := tx.Attempts().FindByID(ctx, attemptID)
		if err != nil {
			return err
		}
		if a.State() == "success" || a.State() == "terminated" {
			return nil
		}
		win, err := tx.Windows().FindByCategory(ctx, a.CategoryID())
		if err != nil {
			return err
		}
		now := uc.engine.Clock().Now()
		if success {
			ar := newAppAttemptRecord(a)
			ar.state = "success"
			ar.version++
			ar.updatedAt = now
			ar.reason = "late success"
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
				res, err := tx.Reservations().FindByID(ctx, ar.reservationID)
				if err != nil {
					return err
				}
				if res.ExpiresAt().Before(win.StartAt()) {
					rr := newAppReservationRecord(res)
					rr.state = "released"
					rr.version++
					rr.updatedAt = now
					if err := tx.Reservations().Save(ctx, rr); err != nil {
						return err
					}
				} else {
					if err := releaseReservationInTx(ctx, tx, uc.engine.Clock(), ar.categoryID, ar.reservationID); err != nil {
						return err
					}
				}
			}
			return nil
		}
		ar := newAppAttemptRecord(a)
		ar.state = "failed"
		ar.failureCount++
		ar.reason = "late failure"
		ar.version++
		ar.updatedAt = now
		if err := tx.Attempts().Save(ctx, ar); err != nil {
			return err
		}
		return tx.Results().Save(ctx, &appResultRecord{
			attemptID:  attemptID,
			success:    false,
			terminated: false,
			recordedAt: now,
			version:    1,
		})
	})
}
