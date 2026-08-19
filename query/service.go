package query

import (
	"context"
	"errors"
	"fmt"
	"time"

	"retryengine/domain"
)

var ErrNotFound = errors.New("query: not found")

// Reader interfaces represent the query read side.
type AttemptReader interface {
	ListAttempts(ctx context.Context) ([]domain.Attempt, error)
	FindAttempt(ctx context.Context, id string) (domain.Attempt, error)
}

type WindowReader interface {
	ListWindows(ctx context.Context) ([]domain.Window, error)
}

type ReservationReader interface {
	ListReservations(ctx context.Context) ([]domain.Reservation, error)
}

type ResultReader interface {
	ListResults(ctx context.Context) ([]domain.Result, error)
}

type BackoffPlanReader interface {
	ListBackoffPlans(ctx context.Context) ([]domain.BackoffPlan, error)
}

// Service implements stable derivation, filtering, sorting and pagination.
type Service struct {
	attempts     AttemptReader
	windows      WindowReader
	reservations ReservationReader
	results      ResultReader
	backoffPlans BackoffPlanReader
}

func NewService(attempts AttemptReader, windows WindowReader, reservations ReservationReader, results ResultReader, backoffPlans BackoffPlanReader) *Service {
	return &Service{
		attempts:     attempts,
		windows:      windows,
		reservations: reservations,
		results:      results,
		backoffPlans: backoffPlans,
	}
}

func NewQueryService(attempts AttemptReader, windows WindowReader, reservations ReservationReader, results ResultReader, backoffPlans BackoffPlanReader) *Service {
	return NewService(attempts, windows, reservations, results, backoffPlans)
}

// ListAttempts filters, sorts and paginates attempts.
func (s *Service) ListAttempts(ctx context.Context, filter Filter, page Page) ([]domain.Attempt, error) {
	attempts, err := s.attempts.ListAttempts(ctx)
	if err != nil {
		return nil, fmt.Errorf("query: list attempts: %w", err)
	}
	filtered := make([]domain.Attempt, 0, len(attempts))
	for _, a := range attempts {
		if filter.MatchAttempt(a) {
			filtered = append(filtered, a)
		}
	}
	filtered = SortAttempts(filtered, SortAsc)
	result := PaginateAttempts(filtered, page)
	return result.Items, nil
}

// RetryableCandidates returns stable retryable attempts with next run, failure count and remaining budget.
func (s *Service) RetryableCandidates(ctx context.Context, filter Filter, page Page) ([]RetryableCandidate, error) {
	attempts, err := s.attempts.ListAttempts(ctx)
	if err != nil {
		return nil, fmt.Errorf("query: list attempts: %w", err)
	}
	windows, err := s.windows.ListWindows(ctx)
	if err != nil {
		return nil, fmt.Errorf("query: list windows: %w", err)
	}
	plans, err := s.backoffPlans.ListBackoffPlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("query: list backoff plans: %w", err)
	}
	results, err := s.results.ListResults(ctx)
	if err != nil {
		return nil, fmt.Errorf("query: list results: %w", err)
	}

	budgetByCategory := make(map[string]int)
	for _, w := range windows {
		remaining := w.Limit() - w.Used() - w.Reserved()
		if remaining < 0 {
			remaining = 0
		}
		budgetByCategory[w.PolicyID()] = remaining
	}
	nextByAttempt := make(map[string]time.Time)
	for _, p := range plans {
		nextByAttempt[p.AttemptID()] = p.NextAttemptAt()
	}
	failureByAttempt := make(map[string]int)
	for _, r := range results {
		failureByAttempt[r.AttemptID()]++
	}

	candidates := make([]RetryableCandidate, 0)
	for _, a := range attempts {
		if !filter.MatchAttempt(a) {
			continue
		}
		next := nextByAttempt[a.ID()]
		if next.IsZero() {
			next = a.CreatedAt()
		}
		candidate := NewRetryableCandidate(a, next, failureByAttempt[a.ID()], budgetByCategory[a.CategoryID()])
		candidates = append(candidates, candidate)
	}

	SortCandidates(candidates)
	return PaginateCandidates(candidates, page), nil
}

// RemainingBudget returns remaining quota for a category.
func (s *Service) RemainingBudget(ctx context.Context, category string) (int, error) {
	windows, err := s.windows.ListWindows(ctx)
	if err != nil {
		return 0, fmt.Errorf("query: list windows: %w", err)
	}
	for _, w := range windows {
		if w.PolicyID() == category {
			remaining := w.Limit() - w.Used() - w.Reserved()
			if remaining < 0 {
				remaining = 0
			}
			return remaining, nil
		}
	}
	return 0, fmt.Errorf("query: category %s: %w", category, ErrNotFound)
}

// FailureTrend returns a time-ordered failure trend for a category.
func (s *Service) FailureTrend(ctx context.Context, category string, since time.Time, limit int) ([]FailurePoint, error) {
	attempts, err := s.attempts.ListAttempts(ctx)
	if err != nil {
		return nil, fmt.Errorf("query: list attempts: %w", err)
	}
	results, err := s.results.ListResults(ctx)
	if err != nil {
		return nil, fmt.Errorf("query: list results: %w", err)
	}
	points := BuildFailureTrend(attempts, results, category)
	filtered := make([]FailurePoint, 0)
	for _, p := range points {
		if p.At.Before(since) {
			continue
		}
		filtered = append(filtered, p)
	}
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered, nil
}

// NextExecutable returns the next execution time for an attempt.
func (s *Service) NextExecutable(ctx context.Context, attemptID string) (time.Time, error) {
	plans, err := s.backoffPlans.ListBackoffPlans(ctx)
	if err != nil {
		return time.Time{}, fmt.Errorf("query: list backoff plans: %w", err)
	}
	for _, p := range plans {
		if p.AttemptID() == attemptID {
			return p.NextAttemptAt(), nil
		}
	}
	return time.Time{}, fmt.Errorf("query: attempt %s: %w", attemptID, ErrNotFound)
}
