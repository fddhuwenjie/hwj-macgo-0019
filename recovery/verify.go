package recovery

import (
	"fmt"

	"retryengine/domain"
)

// VerifySnapshot checks snapshot aggregates for version monotonicity and basic invariants.
func VerifySnapshot(snap Snapshot) error {
	if err := VerifyAttempts(snap.Attempts); err != nil {
		return err
	}
	if err := VerifyWindows(snap.Windows); err != nil {
		return err
	}
	if err := VerifyReservations(snap.Reservations); err != nil {
		return err
	}
	return nil
}

func VerifyAttempts(attempts []domain.Attempt) error {
	seen := make(map[string]int64, len(attempts))
	for _, a := range attempts {
		if a.ID() == "" {
			return fmt.Errorf("recovery: attempt with empty id")
		}
		if a.Version() == 0 {
			return fmt.Errorf("recovery: attempt %s has zero version", a.ID())
		}
		if prev, ok := seen[a.ID()]; ok && a.Version() < prev {
			return fmt.Errorf("recovery: attempt %s version went backwards", a.ID())
		}
		seen[a.ID()] = a.Version()
	}
	return nil
}

func VerifyWindows(windows []domain.Window) error {
	seen := make(map[string]int64, len(windows))
	for _, w := range windows {
		if w.ID() == "" {
			return fmt.Errorf("recovery: window with empty id")
		}
		if w.Limit() < 0 || w.Used() < 0 || w.Reserved() < 0 {
			return fmt.Errorf("recovery: window %s has negative quota/used/reserved", w.ID())
		}
		if w.Used()+w.Reserved() > w.Limit() {
			return fmt.Errorf("recovery: window %s used+reserved exceeds quota", w.ID())
		}
		if prev, ok := seen[w.ID()]; ok && w.Version() < prev {
			return fmt.Errorf("recovery: window %s version went backwards", w.ID())
		}
		seen[w.ID()] = w.Version()
	}
	return nil
}

func VerifyReservations(reservations []domain.Reservation) error {
	for _, r := range reservations {
		if r.ID() == "" {
			return fmt.Errorf("recovery: reservation with empty id")
		}
		if r.Amount() < 0 {
			return fmt.Errorf("recovery: reservation %s has negative quantity", r.ID())
		}
	}
	return nil
}
