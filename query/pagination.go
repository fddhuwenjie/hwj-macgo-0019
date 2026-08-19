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

type PageResult struct {
	Items      []domain.Attempt
	Total      int
	Offset     int
	Limit      int
	NextOffset int
	HasMore    bool
}

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

func PaginateAttempts(items []domain.Attempt, page Page) PageResult {
	if err := ValidatePage(page); err != nil {
		return PageResult{Items: nil, Total: len(items), Offset: page.Offset, Limit: page.Limit}
	}
	total := len(items)
	if page.Offset > total {
		page.Offset = total
	}
	end := page.Offset + page.Limit
	if end > total {
		end = total
	}
	pageItems := append([]domain.Attempt(nil), items[page.Offset:end]...)
	next := end
	hasMore := end < total
	return PageResult{
		Items:      pageItems,
		Total:      total,
		Offset:     page.Offset,
		Limit:      page.Limit,
		NextOffset: next,
		HasMore:    hasMore,
	}
}
