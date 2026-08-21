package journal

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

type Writer struct {
	mu   sync.Mutex
	file *os.File
	path string
	seq  uint64
}

func OpenWriter(path string) (*Writer, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	w := &Writer{file: f, path: path}
	if err := w.recoverSequence(); err != nil {
		_ = f.Close()
		return nil, err
	}
	return w, nil
}

func (w *Writer) recoverSequence() error {
	r, err := OpenReader(w.path)
	if err != nil {
		return err
	}
	recs, err := r.ReadAll(context.Background())
	if err != nil {
		return err
	}
	if len(recs) > 0 {
		w.seq = recs[len(recs)-1].Sequence + 1
	}
	return nil
}

func (w *Writer) Append(ctx context.Context, version int64, payload []byte) (Record, error) {
	select {
	case <-ctx.Done():
		return Record{}, ctx.Err()
	default:
	}
	rec := Record{
		Sequence:  w.seq,
		Version:   version,
		Length:    uint32(len(payload)),
		Payload:   append([]byte(nil), payload...),
		Timestamp: time.Now().UTC().UnixNano(),
	}
	rec.Checksum = rec.computeChecksum()
	raw := rec.Encode()
	if _, err := w.file.Write(raw); err != nil {
		return Record{}, err
	}
	if err := w.file.Sync(); err != nil {
		return Record{}, err
	}
	w.seq++
	return rec, nil
}

func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Sync()
	if closeErr := w.file.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	w.file = nil
	return err
}

func (w *Writer) Path() string {
	return w.path
}

func (w *Writer) String() string {
	return fmt.Sprintf("journal.Writer(%s)", w.path)
}
