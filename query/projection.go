package query

import (
	"sort"
	"time"

	"retryengine/domain"
)

// RetryableCandidate is a derived query result for retry scheduling.
type RetryableCandidate struct {
	Attempt         domain.Attempt `json:"attempt"`
	NextRun         time.Time      `json:"next_run"`
	FailureCount    int            `json:"failure_count"`
	RemainingBudget int            `json:"remaining_budget"`
}

// FailurePoint is a derived point in a failure trend series.
type FailurePoint struct {
	Category string
	At       time.Time
	Count    int
	State    domain.AttemptState
}

func NewRetryableCandidate(attempt domain.Attempt, nextRun time.Time, failureCount int, remainingBudget int) RetryableCandidate {
	return RetryableCandidate{
		Attempt:         attempt,
		NextRun:         nextRun,
		FailureCount:    failureCount,
		RemainingBudget: remainingBudget,
	}
}

func SortCandidates(candidates []RetryableCandidate) {
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].NextRun.Equal(candidates[j].NextRun) {
			return candidates[i].Attempt.ID() < candidates[j].Attempt.ID()
		}
		return candidates[i].NextRun.Before(candidates[j].NextRun)
	})
}

func BuildFailureTrend(attempts []domain.Attempt, results []domain.Result, category string) []FailurePoint {
	attemptCategory := make(map[string]string, len(attempts))
	for _, a := range attempts {
		attemptCategory[a.ID()] = a.CategoryID()
	}
	countByTime := map[time.Time]int{}
	for _, r := range results {
		cat, ok := attemptCategory[r.AttemptID()]
		if !ok || cat != category {
			continue
		}
		countByTime[r.ObservedAt()]++
	}
	points := make([]FailurePoint, 0, len(countByTime))
	for at, count := range countByTime {
		points = append(points, FailurePoint{Category: category, At: at, Count: count})
	}
	sort.Slice(points, func(i, j int) bool {
		return points[i].At.Before(points[j].At)
	})
	return points
}

func PaginateCandidates(candidates []RetryableCandidate, page Page) Paged[RetryableCandidate] {
	// Paginate clamps offset/end into range and preserves Total even when the
	// page is empty, so a stale client offset never panics and metadata stays
	// consistent with the attempt pagination path.
	return Paginate(candidates, page)
}
