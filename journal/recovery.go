package journal

import (
	"context"
	"fmt"
	"os"
)

func TruncateToLastValid(ctx context.Context, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(raw) == 0 {
		return nil
	}
	offset := 0
	lastValid := 0
	for offset < len(raw) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		rec, n, err := DecodeRecord(raw[offset:])
		if err != nil {
			if offset == lastValid {
				// Only a partial or corrupt tail after the last valid record.
				return os.Truncate(path, int64(lastValid))
			}
			return fmt.Errorf("%w: cannot safely truncate", ErrMidFileCorrupt)
		}
		_ = rec
		offset += n
		lastValid = offset
	}
	return nil
}
