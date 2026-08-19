package recovery

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"retryengine/domain"
)

// LogEntry is a parsed write-ahead log record.
type LogEntry struct {
	Sequence  uint64          `json:"sequence"`
	Version   uint64          `json:"version"`
	Length    int             `json:"length"`
	Checksum  string          `json:"checksum"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

// JournalReader reads ordered journal entries.
type JournalReader interface {
	Next(ctx context.Context) (LogEntry, error)
	Close() error
}

// FileJournalReader reads JSONL journal records from a file.
type FileJournalReader struct {
	file *os.File
	dec  *bufio.Reader
}

func OpenFileJournalReader(path string) (*FileJournalReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("recovery: open journal: %w", err)
	}
	return &FileJournalReader{file: f, dec: bufio.NewReader(f)}, nil
}

func (r *FileJournalReader) Next(ctx context.Context) (LogEntry, error) {
	if err := ctx.Err(); err != nil {
		return LogEntry{}, err
	}
	for {
		line, err := r.dec.ReadBytes('\n')
		if err != nil && len(line) == 0 {
			if errors.Is(err, io.EOF) {
				return LogEntry{}, io.EOF
			}
			return LogEntry{}, fmt.Errorf("recovery: read journal: %w", err)
		}
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var entry LogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			return LogEntry{}, fmt.Errorf("recovery: decode journal line: %w", err)
		}
		return entry, nil
	}
}

func (r *FileJournalReader) Close() error {
	return r.file.Close()
}

// RecoveryStore is the persistence boundary used during replay.
type RecoveryStore interface {
	SaveAttempt(ctx context.Context, attempt domain.Attempt) error
	SaveWindow(ctx context.Context, window domain.Window) error
	SaveReservation(ctx context.Context, reservation domain.Reservation) error
	SaveResult(ctx context.Context, result domain.Result) error
	SaveBackoffPlan(ctx context.Context, plan domain.BackoffPlan) error
}

// ReplayResult reports replay progress.
type ReplayResult struct {
	Applied      int
	LastSequence uint64
	LastVersion  uint64
}

// Replayer replays journal entries onto a recovery store.
type Replayer struct {
	store RecoveryStore
}

func NewReplayer(store RecoveryStore) *Replayer {
	return &Replayer{store: store}
}

func (r *Replayer) Apply(ctx context.Context, reader JournalReader) (ReplayResult, error) {
	var res ReplayResult
	for {
		entry, err := reader.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return res, fmt.Errorf("recovery: journal read: %w", err)
		}
		if err := ctx.Err(); err != nil {
			return res, err
		}
		if err := r.applyEntry(ctx, entry); err != nil {
			return res, fmt.Errorf("recovery: apply sequence %d: %w", entry.Sequence, err)
		}
		res.Applied++
		res.LastSequence = entry.Sequence
		res.LastVersion = entry.Version
	}
	return res, nil
}

func (r *Replayer) applyEntry(ctx context.Context, entry LogEntry) error {
	switch entry.Type {
	case "attempt":
		var v domain.Attempt
		if err := json.Unmarshal(entry.Payload, &v); err != nil {
			return fmt.Errorf("decode attempt: %w", err)
		}
		return r.store.SaveAttempt(ctx, v)
	case "window":
		var v domain.Window
		if err := json.Unmarshal(entry.Payload, &v); err != nil {
			return fmt.Errorf("decode window: %w", err)
		}
		return r.store.SaveWindow(ctx, v)
	case "reservation":
		var v domain.Reservation
		if err := json.Unmarshal(entry.Payload, &v); err != nil {
			return fmt.Errorf("decode reservation: %w", err)
		}
		return r.store.SaveReservation(ctx, v)
	case "result":
		var v domain.Result
		if err := json.Unmarshal(entry.Payload, &v); err != nil {
			return fmt.Errorf("decode result: %w", err)
		}
		return r.store.SaveResult(ctx, v)
	case "backoff_plan":
		var v domain.BackoffPlan
		if err := json.Unmarshal(entry.Payload, &v); err != nil {
			return fmt.Errorf("decode backoff plan: %w", err)
		}
		return r.store.SaveBackoffPlan(ctx, v)
	default:
		return fmt.Errorf("unknown journal payload type %q", entry.Type)
	}
}
