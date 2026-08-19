package query

import (
	"sort"

	"retryengine/domain"
)

type SortOrder int

const (
	SortAsc SortOrder = iota
	SortDesc
)

// AttemptSort provides stable sorting with entity id fallback.
type AttemptSort struct {
	items []domain.Attempt
	desc  bool
}

func SortAttempts(items []domain.Attempt, order SortOrder) []domain.Attempt {
	copied := append([]domain.Attempt(nil), items...)
	s := &AttemptSort{items: copied, desc: order == SortDesc}
	sort.Stable(s)
	return copied
}

func (s *AttemptSort) Len() int {
	return len(s.items)
}

func (s *AttemptSort) Swap(i, j int) {
	s.items[i], s.items[j] = s.items[j], s.items[i]
}

func (s *AttemptSort) Less(i, j int) bool {
	a, b := s.items[i], s.items[j]
	if !a.CreatedAt().Equal(b.CreatedAt()) {
		if s.desc {
			return a.CreatedAt().After(b.CreatedAt())
		}
		return a.CreatedAt().Before(b.CreatedAt())
	}
	if a.CategoryID() != b.CategoryID() {
		if s.desc {
			return a.CategoryID() > b.CategoryID()
		}
		return a.CategoryID() < b.CategoryID()
	}
	if a.ID() != b.ID() {
		if s.desc {
			return a.ID() > b.ID()
		}
		return a.ID() < b.ID()
	}
	return a.Version() < b.Version()
}
