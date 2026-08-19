package application

import (
	"context"
	"sort"
	"strings"
)

type QueryUseCase struct {
	engine *Engine
}

func NewQueryUseCase(engine *Engine) *QueryUseCase {
	return &QueryUseCase{engine: engine}
}

func (uc *QueryUseCase) ListCandidates(ctx context.Context, categoryID string, page Page) ([]CandidateView, error) {
	var candidates []CandidateView
	err := withTx(ctx, uc.engine.Store(), func(tx Tx) error {
		attempts, err := tx.Attempts().ListRetryable(ctx, categoryID, uc.engine.Clock().Now(), maxPageLimit(page.Limit))
		if err != nil {
			return err
		}
		for _, a := range attempts {
			candidates = append(candidates, CandidateView{
				AttemptID:     a.ID(),
				CategoryID:    a.CategoryID(),
				NextAttemptAt: a.NextAttemptAt(),
				FailureCount:  a.FailureCount(),
				Reason:        a.Reason(),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sortCandidates(candidates, page)
	return paginateCandidates(candidates, page.Offset, page.Limit), nil
}

func maxPageLimit(limit int) int {
	if limit <= 0 || limit > 100 {
		return 100
	}
	return limit
}

func sortCandidates(items []CandidateView, page Page) {
	less := func(i, j int) bool {
		if page.SortBy == "failure_count" {
			if items[i].FailureCount != items[j].FailureCount {
				return items[i].FailureCount > items[j].FailureCount
			}
		} else {
			if !items[i].NextAttemptAt.Equal(items[j].NextAttemptAt) {
				return items[i].NextAttemptAt.Before(items[j].NextAttemptAt)
			}
		}
		return strings.Compare(items[i].AttemptID, items[j].AttemptID) < 0
	}
	if page.Descending {
		old := less
		less = func(i, j int) bool { return old(j, i) }
	}
	sort.SliceStable(items, less)
}

func paginateCandidates(items []CandidateView, offset, limit int) []CandidateView {
	if offset < 0 {
		offset = 0
	}
	if offset > len(items) {
		return []CandidateView{}
	}
	end := len(items)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return items[offset:end]
}
