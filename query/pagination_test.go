package query

import (
	"testing"
	"time"

	"retryengine/domain"
)

var baseTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// makeAttempts builds n attempts with stable, distinguishable ids and
// monotonically increasing creation times so sorting is deterministic.
func makeAttempts(n int) []domain.Attempt {
	out := make([]domain.Attempt, n)
	for i := 0; i < n; i++ {
		a, err := domain.NewAttempt(
			"att-"+string(rune('a'+i)),
			"cat", "pol", "win", "res",
			i,
			baseTime.Add(time.Duration(i)*time.Second),
		)
		if err != nil {
			panic("makeAttempts: " + err.Error())
		}
		out[i] = *a
	}
	return out
}

// makeCandidates builds n retryable candidates with deterministic next-run
// times.
func makeCandidates(n int) []RetryableCandidate {
	out := make([]RetryableCandidate, n)
	for i := 0; i < n; i++ {
		out[i] = NewRetryableCandidate(
			domain.Attempt{}, baseTime.Add(time.Duration(i)*time.Second), i, 100-i,
		)
	}
	return out
}

func TestPaginateAttempts_OffsetBeyondTotalReturnsEmptyPageWithCorrectTotal(t *testing.T) {
	// Simulate a list refresh shrinking the data: the client still holds a
	// stale offset past the new end.
	items := makeAttempts(3)
	page := Page{Offset: 10, Limit: 5}

	got := PaginateAttempts(items, page)

	if len(got.Items) != 0 {
		t.Fatalf("expected empty page, got %d items", len(got.Items))
	}
	if got.Total != 3 {
		t.Fatalf("expected total=3 (preserved), got %d", got.Total)
	}
	if got.HasMore {
		t.Fatalf("expected HasMore=false, got true")
	}
	if got.NextOffset != 3 {
		t.Fatalf("expected NextOffset=3, got %d", got.NextOffset)
	}
}

func TestPaginateAttempts_OffsetBeyondTotalDoesNotPanic(t *testing.T) {
	// Previously items[Offset:end] panicked here. This test fails (via panic)
	// if the regression returns.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("PaginateAttempts panicked on out-of-range offset: %v", r)
		}
	}()

	items := makeAttempts(2)
	_ = PaginateAttempts(items, Page{Offset: 1_000_000, Limit: 10})
}

func TestPaginateAttempts_OffsetEqualsLengthReturnsEmptyPage(t *testing.T) {
	items := makeAttempts(4)
	// Offset exactly at the boundary is the canonical "no more rows" cursor.
	got := PaginateAttempts(items, Page{Offset: 4, Limit: 5})

	if len(got.Items) != 0 {
		t.Fatalf("expected empty page at exact boundary, got %d items", len(got.Items))
	}
	if got.Total != 4 {
		t.Fatalf("expected total=4, got %d", got.Total)
	}
	if got.HasMore {
		t.Fatalf("expected HasMore=false, got true")
	}
}

func TestPaginateAttempts_LastPageShorterThanLimitReturnsRemaining(t *testing.T) {
	items := makeAttempts(7)
	// 7 items, offset 5, limit 5 -> only 2 remain on the last page.
	got := PaginateAttempts(items, Page{Offset: 5, Limit: 5})

	if len(got.Items) != 2 {
		t.Fatalf("expected 2 remaining items on last page, got %d", len(got.Items))
	}
	if got.Total != 7 {
		t.Fatalf("expected total=7, got %d", got.Total)
	}
	if got.HasMore {
		t.Fatalf("expected HasMore=false on last page, got true")
	}
	if got.NextOffset != 7 {
		t.Fatalf("expected NextOffset=7, got %d", got.NextOffset)
	}
}

func TestPaginateAttempts_FullPageReportsHasMore(t *testing.T) {
	items := makeAttempts(10)
	got := PaginateAttempts(items, Page{Offset: 0, Limit: 5})

	if len(got.Items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(got.Items))
	}
	if !got.HasMore {
		t.Fatalf("expected HasMore=true, got false")
	}
	if got.NextOffset != 5 {
		t.Fatalf("expected NextOffset=5, got %d", got.NextOffset)
	}
	if got.Total != 10 {
		t.Fatalf("expected total=10, got %d", got.Total)
	}
}

func TestPaginateAttempts_ReturnsACopy(t *testing.T) {
	items := makeAttempts(3)
	got := PaginateAttempts(items, Page{Offset: 0, Limit: 3})

	// Mutating the result must not affect the source slice.
	if len(got.Items) == 0 {
		t.Fatal("expected non-empty page")
	}
	originalID := items[0].ID()
	got.Items[0] = domain.Attempt{}
	if items[0].ID() != originalID {
		t.Fatalf("PaginateAttempts returned a slice aliasing the source; source mutated to %q", items[0].ID())
	}
}

func TestPaginateAttempts_NegativeOffsetIsClampedToZero(t *testing.T) {
	items := makeAttempts(3)
	got := PaginateAttempts(items, Page{Offset: -5, Limit: 2})

	if len(got.Items) != 2 {
		t.Fatalf("expected 2 items after clamping negative offset, got %d", len(got.Items))
	}
	if got.Offset != 0 {
		t.Fatalf("expected clamped Offset=0, got %d", got.Offset)
	}
}

func TestPaginateCandidates_OffsetBeyondTotalReturnsEmptyPageWithCorrectTotal(t *testing.T) {
	candidates := makeCandidates(3)
	got := PaginateCandidates(candidates, Page{Offset: 10, Limit: 5})

	if len(got.Items) != 0 {
		t.Fatalf("expected empty page, got %d items", len(got.Items))
	}
	if got.Total != 3 {
		t.Fatalf("expected total=3, got %d", got.Total)
	}
	if got.HasMore {
		t.Fatalf("expected HasMore=false, got true")
	}
}

func TestPaginateCandidates_LastPageShorterThanLimitReturnsRemaining(t *testing.T) {
	candidates := makeCandidates(7)
	got := PaginateCandidates(candidates, Page{Offset: 5, Limit: 5})

	if len(got.Items) != 2 {
		t.Fatalf("expected 2 remaining items, got %d", len(got.Items))
	}
	if got.Total != 7 {
		t.Fatalf("expected total=7, got %d", got.Total)
	}
	if got.HasMore {
		t.Fatalf("expected HasMore=false on last page, got true")
	}
}

func TestPaginateCandidates_MetadataMatchesAttemptsPath(t *testing.T) {
	// The two pagination paths must produce structurally identical
	// metadata for the same page geometry.
	attempts := makeAttempts(6)
	candidates := makeCandidates(6)
	page := Page{Offset: 4, Limit: 5}

	a := PaginateAttempts(attempts, page)
	c := PaginateCandidates(candidates, page)

	if a.Total != c.Total || a.Offset != c.Offset || a.Limit != c.Limit ||
		a.NextOffset != c.NextOffset || a.HasMore != c.HasMore ||
		len(a.Items) != len(c.Items) {
		t.Fatalf("metadata drift between paths:\n  attempts = %+v\n  candidates = %+v", a, c)
	}
}

func TestValidatePage(t *testing.T) {
	cases := []struct {
		name    string
		page    Page
		wantErr bool
	}{
		{"valid", Page{Offset: 0, Limit: 10}, false},
		{"zero offset", Page{Offset: 0, Limit: 1}, false},
		{"negative offset", Page{Offset: -1, Limit: 10}, true},
		{"negative limit", Page{Offset: 0, Limit: -1}, true},
		{"zero limit", Page{Offset: 0, Limit: 0}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePage(tc.page)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidatePage(%+v) err=%v, wantErr=%v", tc.page, err, tc.wantErr)
			}
		})
	}
}

