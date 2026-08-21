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

func (uc *QueryUseCase) ListCandidates(ctx context.Context, categoryID string, page Page) (CandidatePage, error) {
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
		return CandidatePage{}, err
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

// paginateCandidates slices a sorted list into a single page with consistent
// pagination metadata.
//
// Boundary handling (kept identical to the query package's contract):
//   - An offset at or beyond the end of the list yields an empty page (e.g.
//     when a list refresh shrinks the data and the client still holds a stale
//     offset). Total still reflects the full sorted size.
//   - When fewer than limit items remain on the last page, only the remaining
//     items are returned.
//   - Offset and the slice end are clamped to the available range, so the
//     operation never panics on out-of-range input.
func paginateCandidates(items []CandidateView, offset, limit int) CandidatePage {
	total := len(items)

	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}

	end := total
	if limit > 0 {
		end = offset + limit
		if end > total {
			end = total
		}
	}

	var pageItems []CandidateView
	if offset < end {
		// Copy so the caller cannot mutate the underlying slice through a
		// shared backing array.
		pageItems = append([]CandidateView(nil), items[offset:end]...)
	}

	next := end
	hasMore := end < total

	return CandidatePage{
		Items:      pageItems,
		Total:      total,
		Offset:     offset,
		Limit:      limit,
		NextOffset: next,
		HasMore:    hasMore,
	}
}
