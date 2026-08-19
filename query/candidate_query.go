package query

import (
	"sort"
	"strings"
)

// Candidate is a stable retry candidate projection used by query services.
type Candidate struct {
	ID          string
	Category    string
	NextAttempt int64
	Score       int
}

// StableCandidateQuery filters and ranks retry candidates.
type StableCandidateQuery struct {
	MaxResults int
}

// NewStableCandidateQuery returns a query with a safe default page size.
func NewStableCandidateQuery(maxResults int) StableCandidateQuery {
	if maxResults <= 0 {
		maxResults = 50
	}
	return StableCandidateQuery{MaxResults: maxResults}
}

// SortCandidates returns a stable copy of candidates sorted by category asc,
// next attempt asc, and ID asc as a tie breaker.
func (q StableCandidateQuery) SortCandidates(in []Candidate) []Candidate {
	out := append([]Candidate(nil), in...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Category != out[j].Category {
			return out[i].Category < out[j].Category
		}
		if out[i].NextAttempt != out[j].NextAttempt {
			return out[i].NextAttempt < out[j].NextAttempt
		}
		return strings.Compare(out[i].ID, out[j].ID) < 0
	})
	if len(out) > q.MaxResults {
		out = out[:q.MaxResults]
	}
	return out
}
