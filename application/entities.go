package application

import "time"

type BudgetPolicy struct {
	WindowDuration  time.Duration
	MaxAttempts     int
	MaxReservations int
	BackoffBase     time.Duration
	BackoffMax      time.Duration
}

type AttemptCommand struct {
	IdempotencyKey string
	CategoryID     string
	ReservationID  string
}

type AttemptOutcome struct {
	AttemptID  string
	Success    bool
	Terminated bool
	Reason     string
}

type BatchResult struct {
	Succeeded int
	Failed    int
	Items     []ItemResult
}

type ItemResult struct {
	Index   int
	Key     string
	Success bool
	Error   string
}

type BudgetView struct {
	CategoryID          string
	Used                int
	Reserved            int
	Remaining           int
	WindowStart         time.Time
	WindowEnd           time.Time
	ConsecutiveFailures int
}

type WindowView struct {
	CategoryID string
	Used       int
	Reserved   int
	StartAt    time.Time
	EndAt      time.Time
}

type CandidateView struct {
	AttemptID     string
	CategoryID    string
	NextAttemptAt time.Time
	FailureCount  int
	Reason        string
}

// CandidatePage is the pagination envelope for candidate views. Total is the
// number of candidates after filtering and sorting (independent of the page),
// so the caller can keep the full size even when the page is empty.
type CandidatePage struct {
	Items      []CandidateView
	Total      int
	Offset     int
	Limit      int
	NextOffset int
	HasMore    bool
}

type Page struct {
	Offset     int
	Limit      int
	SortBy     string
	Descending bool
}
