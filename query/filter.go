package query

import (
	"time"

	"retryengine/domain"
)

// Filter is a stable filter applied before sorting and pagination.
type Filter struct {
	Categories        []string
	States            []domain.AttemptState
	WindowID          string
	ReservationID     string
	Since             *time.Time
	Before            *time.Time
	IncludeTerminated bool
}

func (f Filter) MatchAttempt(a domain.Attempt) bool {
	if len(f.Categories) > 0 {
		found := false
		for _, c := range f.Categories {
			if a.CategoryID() == c {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(f.States) > 0 {
		found := false
		for _, st := range f.States {
			if a.State() == st {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if f.WindowID != "" && a.WindowID() != f.WindowID {
		return false
	}
	if f.ReservationID != "" && a.ReservationID() != f.ReservationID {
		return false
	}
	if f.Since != nil && a.CreatedAt().Before(*f.Since) {
		return false
	}
	if f.Before != nil && a.CreatedAt().After(*f.Before) {
		return false
	}
	return true
}
