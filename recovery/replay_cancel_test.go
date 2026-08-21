package recovery

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"retryengine/domain"
)

// writeJournal writes JSONL log entries to path in order.
func writeJournal(t *testing.T, path string, entries []LogEntry) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create journal: %v", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, e := range entries {
		b, err := json.Marshal(e)
		if err != nil {
			t.Fatalf("marshal entry: %v", err)
		}
		if _, err := w.Write(append(b, '\n')); err != nil {
			t.Fatalf("write entry: %v", err)
		}
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush journal: %v", err)
	}
}

// recordingStore is a RecoveryStore that records each saved record type. The
// Replayer passes its ctx through to every Save* method, so cancellation of
// that ctx is observable here too; the chain is tested via reader.Next and
// Apply's ctx.Err() checks first.
type recordingStore struct {
	saved []string
}

func (s *recordingStore) SaveAttempt(context.Context, domain.Attempt) error {
	s.saved = append(s.saved, "attempt")
	return nil
}
func (s *recordingStore) SaveWindow(context.Context, domain.Window) error {
	s.saved = append(s.saved, "window")
	return nil
}
func (s *recordingStore) SaveReservation(context.Context, domain.Reservation) error {
	s.saved = append(s.saved, "reservation")
	return nil
}
func (s *recordingStore) SaveResult(context.Context, domain.Result) error {
	s.saved = append(s.saved, "result")
	return nil
}
func (s *recordingStore) SaveBackoffPlan(context.Context, domain.BackoffPlan) error {
	s.saved = append(s.saved, "backoff_plan")
	return nil
}

// entriesN builds n log entries of mixed types with monotonically increasing
// sequence and version. The empty-object payload decodes into the domain
// types without error (their fields are unexported), which is enough to drive
// the replay control flow under test.
func entriesN(n int) []LogEntry {
	types := []string{"attempt", "window", "reservation", "result", "backoff_plan"}
	out := make([]LogEntry, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, LogEntry{
			Sequence:  uint64(i + 1),
			Version:   uint64(i + 1),
			Length:    2,
			Checksum:  "ok",
			Type:      types[i%len(types)],
			Payload:   json.RawMessage(`{}`),
			CreatedAt: time.Unix(int64(i), 0).UTC(),
		})
	}
	return out
}

// With a live context, Apply must replay every entry and return a complete
// ReplayResult. This pins the non-cancelled behaviour.
func TestReplayerApply_NoCancellationReplaysAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.jsonl")
	writeJournal(t, path, entriesN(5))

	reader, err := OpenFileJournalReader(path)
	if err != nil {
		t.Fatalf("open reader: %v", err)
	}
	defer reader.Close()

	store := &recordingStore{}
	res, err := NewReplayer(store).Apply(context.Background(), reader)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if res.Applied != 5 {
		t.Fatalf("Applied: got %d want 5", res.Applied)
	}
	if res.LastSequence != 5 {
		t.Fatalf("LastSequence: got %d want 5", res.LastSequence)
	}
	if res.LastVersion != 5 {
		t.Fatalf("LastVersion: got %d want 5", res.LastVersion)
	}
	if len(store.saved) != 5 {
		t.Fatalf("store saved: got %d want 5", len(store.saved))
	}
}

// With a cancelled context, Apply must stop early and surface the cancellation
// cause, never returning a clean ReplayResult as if replay completed normally.
func TestReplayerApply_CancelledStopsAndSurfacesCause(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.jsonl")
	writeJournal(t, path, entriesN(5))

	reader, err := OpenFileJournalReader(path)
	if err != nil {
		t.Fatalf("open reader: %v", err)
	}
	defer reader.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	store := &recordingStore{}
	res, err := NewReplayer(store).Apply(ctx, reader)
	if err == nil {
		t.Fatal("expected cancellation error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled in error chain, got %v", err)
	}
	// It must not report a full, normally-completed result.
	if res.Applied == 5 {
		t.Fatalf("Apply reported all %d entries applied despite cancellation", res.Applied)
	}
	if len(store.saved) == 5 {
		t.Fatalf("store saved all %d entries despite cancellation", len(store.saved))
	}
}

// A cancelled FileJournalReader.Next must return the cancellation cause rather
// than blocking or returning EOF.
func TestFileJournalReader_NextCancelledSurfacesCause(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.jsonl")
	writeJournal(t, path, entriesN(2))

	reader, err := OpenFileJournalReader(path)
	if err != nil {
		t.Fatalf("open reader: %v", err)
	}
	defer reader.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = reader.Next(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

// Manager.Recover end-to-end: cancellation must propagate through the whole
// recovery chain, returning the cause and not a fully-recovered snapshot.
func TestManagerRecover_CancelledPropagates(t *testing.T) {
	dir := t.TempDir()
	snapDir := filepath.Join(dir, "snap")
	logPath := filepath.Join(dir, "journal.jsonl")
	writeJournal(t, logPath, entriesN(5))

	store := &recordingStore{}
	mgr := NewManager(snapDir, logPath, store)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	snap, err := mgr.Recover(ctx)
	if err == nil {
		t.Fatal("expected cancellation error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled in error chain, got %v", err)
	}
	// The snapshot must not carry a completed journal sequence.
	if snap.JournalSequence == 5 {
		t.Fatalf("Recover reported completed journal sequence %d despite cancellation", snap.JournalSequence)
	}
	// Recovery must not have fully applied the journal.
	if len(store.saved) == 5 {
		t.Fatalf("store saved all %d entries despite cancellation", len(store.saved))
	}
}

// Manager.Recover end-to-end with a live context replays the full journal and
// advances the snapshot's journal sequence, confirming the non-cancelled result
// is unchanged by the cancellation path.
func TestManagerRecover_NoCancellationReplaysAll(t *testing.T) {
	dir := t.TempDir()
	snapDir := filepath.Join(dir, "snap")
	logPath := filepath.Join(dir, "journal.jsonl")
	writeJournal(t, logPath, entriesN(4))

	store := &recordingStore{}
	mgr := NewManager(snapDir, logPath, store)

	snap, err := mgr.Recover(context.Background())
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if snap.JournalSequence != 4 {
		t.Fatalf("JournalSequence: got %d want 4", snap.JournalSequence)
	}
	if snap.Version != 4 {
		t.Fatalf("Version: got %d want 4", snap.Version)
	}
	if len(store.saved) != 4 {
		t.Fatalf("store saved: got %d want 4", len(store.saved))
	}
}
