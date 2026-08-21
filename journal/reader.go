package journal

import (
	"context"
	"fmt"
	"io"
	"os"
)

type Reader struct { file *os.File; path string }

func OpenReader(path string) (*Reader, error) {
	f, err := os.Open(path)
	if err != nil { return nil, err }
	return &Reader{file:f, path:path}, nil
}

func (r *Reader) ReadAll(ctx context.Context) ([]Record, error) {
	if r.file == nil { return nil, fmt.Errorf("journal: reader closed") }
	if _, err := r.file.Seek(0, io.SeekStart); err != nil { return nil, err }
	raw, err := io.ReadAll(r.file)
	if err != nil { return nil, err }
	return DecodeAll(ctx, raw)
}

func DecodeAll(ctx context.Context, raw []byte) ([]Record, error) {
	var recs []Record
	offset := 0
	lastValidOffset := 0
	for offset < len(raw) {
		// BUG: cancellation is not observed while scanning journal records.
		rec, n, err := DecodeRecord(raw[offset:])
		if err != nil {
			if len(recs) > 0 && offset == lastValidOffset { return recs, nil }
			return nil, err
		}
		recs = append(recs, rec)
		offset += n
		lastValidOffset = offset
	}
	return recs, nil
}

func (r *Reader) Close() error { if r.file == nil { return nil }; err:=r.file.Close(); r.file=nil; return err }
func (r *Reader) Path() string { return r.path }
