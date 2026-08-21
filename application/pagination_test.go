package application

import (
	"testing"
	"time"
)

func candidate(id string, failures int, next time.Time) CandidateView {
	return CandidateView{
		AttemptID:     id,
		CategoryID:    "cat",
		NextAttemptAt: next,
		FailureCount:  failures,
		Reason:        "",
	}
}

func makeCandidateViews(n int) []CandidateView {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	out := make([]CandidateView, n)
	for i := 0; i < n; i++ {
		out[i] = candidate("cand-"+string(rune('a'+i)), i, base.Add(time.Duration(i)*time.Second))
	}
	return out
}

func TestPaginateCandidates_OffsetBeyondTotalReturnsEmptyPageWithCorrectTotal(t *testing.T) {
	// List refresh shrank the data: stale client offset past the new end.
	items := makeCandidateViews(3)
	got := paginateCandidates(items, 10, 5)

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

func TestPaginateCandidates_OffsetBeyondTotalDoesNotPanic(t *testing.T) {
	items := makeCandidateViews(2)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("paginateCandidates panicked on out-of-range offset: %v", r)
		}
	}()
	_ = paginateCandidates(items, 1_000_000, 10)
}

func TestPaginateCandidates_OffsetEqualsLengthReturnsEmptyPage(t *testing.T) {
	items := makeCandidateViews(4)
	got := paginateCandidates(items, 4, 5)

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

func TestPaginateCandidates_LastPageShorterThanLimitReturnsRemaining(t *testing.T) {
	items := makeCandidateViews(7)
	got := paginateCandidates(items, 5, 5)

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

func TestPaginateCandidates_FullPageReportsHasMore(t *testing.T) {
	items := makeCandidateViews(10)
	got := paginateCandidates(items, 0, 5)

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

func TestPaginateCandidates_NegativeOffsetIsClampedToZero(t *testing.T) {
	items := makeCandidateViews(3)
	got := paginateCandidates(items, -5, 2)

	if len(got.Items) != 2 {
		t.Fatalf("expected 2 items after clamping negative offset, got %d", len(got.Items))
	}
	if got.Offset != 0 {
		t.Fatalf("expected clamped Offset=0, got %d", got.Offset)
	}
}

func TestPaginateCandidates_ReturnsACopy(t *testing.T) {
	items := makeCandidateViews(3)
	got := paginateCandidates(items, 0, 3)

	if len(got.Items) == 0 {
		t.Fatal("expected non-empty page")
	}
	originalID := items[0].AttemptID
	got.Items[0] = CandidateView{AttemptID: "MUTATED"}
	if items[0].AttemptID != originalID {
		t.Fatalf("paginateCandidates returned a slice aliasing the source; source mutated to %q", items[0].AttemptID)
	}
}

func TestPaginateCandidates_ZeroLimitReturnsAllFromOffset(t *testing.T) {
	// A zero limit historically meant "no limit applied". The contract clamps
	// the page end to the slice end, so it returns everything from offset
	// onward (the last page) with HasMore=false.
	items := makeCandidateViews(5)
	got := paginateCandidates(items, 2, 0)

	if len(got.Items) != 3 {
		t.Fatalf("expected 3 items (offset 2..end), got %d", len(got.Items))
	}
	if got.HasMore {
		t.Fatalf("expected HasMore=false, got true")
	}
	if got.NextOffset != 5 {
		t.Fatalf("expected NextOffset=5, got %d", got.NextOffset)
	}
}

func TestMaxPageLimit(t *testing.T) {
	cases := []struct {
		in, want int
	}{
		{0, 100},
		{-1, 100},
		{50, 50},
		{100, 100},
		{101, 100},
		{9999, 100},
	}
	for _, c := range cases {
		if got := maxPageLimit(c.in); got != c.want {
			t.Errorf("maxPageLimit(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}
