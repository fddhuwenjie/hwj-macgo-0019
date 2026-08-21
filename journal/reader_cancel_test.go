package journal

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// encodeN appends n valid records to a buffer and returns it plus the records.
func encodeN(t *testing.T, n int) ([]byte, []Record) {
	t.Helper()
	out := make([]byte, 0)
	recs := make([]Record, 0, n)
	for i := 0; i < n; i++ {
		rec := Record{
			Sequence:  uint64(i),
			Version:   int64(i + 1),
			Length:    uint32(3),
			Payload:   []byte("pay"),
			Timestamp: int64(i),
		}
		rec.Checksum = rec.computeChecksum()
		out = append(out, rec.Encode()...)
		recs = append(recs, rec)
	}
	return out, recs
}

// DecodeAll with a live context must return every record unchanged. This pins
// the non-cancelled behaviour so the cancellation path can be added without
// altering normal results.
func TestDecodeAll_NoCancellationReturnsAllRecords(t *testing.T) {
	raw, want := encodeN(t, 5)

	got, err := DecodeAll(context.Background(), raw)
	if err != nil {
		t.Fatalf("DecodeAll returned error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("record count: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Sequence != want[i].Sequence {
			t.Fatalf("record %d sequence: got %d want %d", i, got[i].Sequence, want[i].Sequence)
		}
		if got[i].Version != want[i].Version {
			t.Fatalf("record %d version: got %d want %d", i, got[i].Version, want[i].Version)
		}
		if string(got[i].Payload) != string(want[i].Payload) {
			t.Fatalf("record %d payload: got %q want %q", i, got[i].Payload, want[i].Payload)
		}
	}
}

// DecodeAll with a pre-cancelled context must stop before scanning and surface
// the cancellation cause, never returning a partial-but-normal record slice.
func TestDecodeAll_CancelledStopsAndSurfacesCause(t *testing.T) {
	raw, _ := encodeN(t, 5)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got, err := DecodeAll(ctx, raw)
	if got != nil {
		t.Fatalf("expected nil records on cancellation, got %d", len(got))
	}
	if err == nil {
		t.Fatal("expected cancellation error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled in error chain, got %v", err)
	}
}

// ReadAll must honour a cancelled context before performing any I/O.
func TestReadAll_CancelledStopsBeforeIO(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.bin")
	raw, _ := encodeN(t, 3)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write journal: %v", err)
	}

	r, err := OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	defer r.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got, err := r.ReadAll(ctx)
	if got != nil {
		t.Fatalf("expected nil records on cancellation, got %d", len(got))
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled in error chain, got %v", err)
	}
}

// ReadAll with a live context must return the same records as a fresh decode,
// confirming the cancellation fast-path did not alter the normal result.
func TestReadAll_NoCancellationMatchesDecode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.bin")
	raw, want := encodeN(t, 4)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write journal: %v", err)
	}

	r, err := OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	defer r.Close()

	got, err := r.ReadAll(context.Background())
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("record count: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Sequence != want[i].Sequence {
			t.Fatalf("record %d sequence: got %d want %d", i, got[i].Sequence, want[i].Sequence)
		}
	}

	// A closed reader must report its state instead of performing I/O.
	if err := r.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := r.ReadAll(context.Background()); err == nil {
		t.Fatal("expected error reading closed reader")
	}
}
