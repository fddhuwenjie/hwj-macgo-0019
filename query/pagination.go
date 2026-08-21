package query

import (
	"errors"

	"retryengine/domain"
)

var ErrInvalidPage = errors.New("query: invalid page")

// Page describes offset and limit for pagination.
type Page struct {
	Offset int
	Limit  int
}

// Paged is the canonical pagination envelope shared by every query path.
// Total is the number of items after filtering and sorting (independent of
// the requested page), so callers keep the full dataset size even when the
// page itself is empty. Offset is the clamped offset actually applied.
type Paged[T any] struct {
	Items      []T
	Total      int
	Offset     int
	Limit      int
	NextOffset int
	HasMore    bool
}

// PageResult is the attempt-specific view of Paged, kept for readability at
// call sites that deal only in attempts.
type PageResult = Paged[domain.Attempt]

func ValidatePage(page Page) error {
	if page.Offset < 0 {
		return ErrInvalidPage
	}
	if page.Limit < 0 {
		return ErrInvalidPage
	}
	if page.Limit == 0 {
		return ErrInvalidPage
	}
	return nil
}

// Paginate is the single source of truth for slicing a sorted, filtered list
// into one page. It is panic-free on out-of-range input and keeps the
// pagination metadata consistent across all query paths.
//
// Boundary handling:
//   - An offset at or beyond the end of the list yields an empty page (e.g.
//     when a list refresh shrinks the data and the client still holds a stale
//     offset). Total still reflects the full filtered size so the caller can
//     reconcile its client-side cursor.
//   - When fewer than Limit items remain on the last page, only the remaining
//     items are returned; NextOffset and HasMore describe the (lack of a)
//     following page.
//   - Offset and the slice end are clamped to the available range, so the
//     operation never panics on out-of-range input.
func Paginate[T any](items []T, page Page) Paged[T] {
	total := len(items)

	// Clamp the offset into the valid range. An offset past the end maps to an
	// empty page rather than a panic; the total is preserved.
	offset := page.Offset
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}

	end := offset + page.Limit
	if end > total {
		end = total
	}

	var pageItems []T
	if offset < end {
		pageItems = append([]T(nil), items[offset:end]...)
	}

	next := end
	hasMore := end < total

	return Paged[T]{
		Items:      pageItems,
		Total:      total,
		Offset:     offset,
		Limit:      page.Limit,
		NextOffset: next,
		HasMore:    hasMore,
	}
}

// PaginateAttempts slices a sorted, filtered list of attempts into a single
// page. See Paginate for the boundary contract.
func PaginateAttempts(items []domain.Attempt, page Page) PageResult {
	return Paginate(items, page)
}
