package domain

import (
	"reflect"
	"testing"
	"time"
)

// TestResultMetadataIsolationAfterSave verifies that once a Result is constructed
// (its content "saved" into the aggregate), the caller reusing and mutating its
// own metadata buffer does not change the already-committed Result. Before the
// fix, NewResult stored the caller's map reference, so later mutations leaked
// into the saved result and even surfaced values that were never committed.
func TestResultMetadataIsolationAfterSave(t *testing.T) {
	now := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)
	buf := map[string]string{"region": "us-east", "attempt": "1"}

	r, err := NewResult("res-1", "att-1", "ok", true, "", now, buf, now)
	if err != nil {
		t.Fatalf("NewResult: %v", err)
	}

	// Caller continues to reuse and mutate its own buffer after the save.
	buf["region"] = "eu-west"
	buf["uncommitted"] = "should-not-leak"

	got := r.Metadata()
	if got["region"] != "us-east" {
		t.Fatalf("saved metadata mutated by caller buffer: region=%q want %q", got["region"], "us-east")
	}
	if _, ok := got["uncommitted"]; ok {
		t.Fatalf("saved metadata exposes a value that was never committed: %q", got["uncommitted"])
	}
}

// TestResultMetadataIsolationOnMutationFromOutside confirms that mutating the map
// returned by Metadata() does not affect the Result's internal state, and that the
// read-only view reflects the originally committed values on every call.
func TestResultMetadataIsolationOnMutationFromOutside(t *testing.T) {
	now := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)
	buf := map[string]string{"k": "v"}

	r, err := NewResult("res-2", "att-2", "ok", true, "", now, buf, now)
	if err != nil {
		t.Fatalf("NewResult: %v", err)
	}

	view := r.Metadata()
	view["injected"] = "x"

	again := r.Metadata()
	if _, ok := again["injected"]; ok {
		t.Fatalf("mutating a returned Metadata() view leaked into internal state: %v", again)
	}
	if again["k"] != "v" {
		t.Fatalf("read-only metadata drifted from committed value: k=%q want %q", again["k"], "v")
	}
}

// TestResultMetadataCloneIsolation ensures a cloned Result owns its own metadata
// map, so mutations on the original (or its buffer) do not bleed into the clone.
func TestResultMetadataCloneIsolation(t *testing.T) {
	now := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)
	buf := map[string]string{"region": "us-east"}

	r, err := NewResult("res-3", "att-3", "ok", true, "", now, buf, now)
	if err != nil {
		t.Fatalf("NewResult: %v", err)
	}

	clone := r.Clone()
	if clone == r {
		t.Fatal("Clone returned the same pointer")
	}

	// Mutate the original's internal map via the constructor's buffer reuse and
	// via a fresh mutation path; the clone must stay independent.
	buf["region"] = "eu-west"
	buf["leak"] = "x"

	if clone.Metadata()["region"] != "us-east" {
		t.Fatalf("clone metadata aliased original buffer: region=%q want %q", clone.Metadata()["region"], "us-east")
	}
	if _, ok := clone.Metadata()["leak"]; ok {
		t.Fatalf("clone metadata absorbed an uncommitted value: %v", clone.Metadata())
	}

	// Mutating the clone's metadata view must not touch the original.
	cloneView := clone.Metadata()
	cloneView["from-clone"] = "x"
	if _, ok := r.Metadata()["from-clone"]; ok {
		t.Fatalf("original metadata mutated through clone: %v", r.Metadata())
	}
}

// TestResultOtherReadOnlyFieldsPreserved confirms the fields other than metadata
// (attemptID, success, failureReason, outcome, observedAt, and the base
// aggregate's id/version/createdAt/updatedAt) remain correct and read-only after
// the metadata-isolation fix.
func TestResultOtherReadOnlyFieldsPreserved(t *testing.T) {
	now := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)
	observed := now.Add(5 * time.Second)
	buf := map[string]string{"k": "v"}

	r, err := NewResult("res-4", "att-4", "done", false, "timeout", observed, buf, now)
	if err != nil {
		t.Fatalf("NewResult: %v", err)
	}

	cases := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"AttemptID", r.AttemptID(), "att-4"},
		{"Success", r.Success(), false},
		{"FailureReason", r.FailureReason(), "timeout"},
		{"Outcome", r.Outcome(), "done"},
		{"ObservedAt", r.ObservedAt(), observed},
		{"ID", r.ID(), "res-4"},
		{"Version", r.Version(), int64(1)},
		{"CreatedAt", r.CreatedAt(), now},
		{"UpdatedAt", r.UpdatedAt(), now},
	}
	for _, c := range cases {
		if !reflect.DeepEqual(c.got, c.want) {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}

	// LateFor(deadline) reports whether the result was observed after the
	// deadline (i.e. arrived late). It is a derived read-only computation and
	// must keep working after the metadata-isolation fix.
	if r.LateFor(observed.Add(time.Second)) {
		t.Errorf("LateFor should be false when deadline is after observedAt")
	}
	if !r.LateFor(observed.Add(-time.Second)) {
		t.Errorf("LateFor should be true when deadline is before observedAt")
	}
}
