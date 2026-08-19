package clockidcodec

import (
	"fmt"
	"strings"
	"time"
)

const (
	MaxBatchSize       = 1000
	DefaultPageSize    = 50
	MinWindowDuration  = time.Millisecond
	MaxBackoffDuration = 24 * time.Hour
	MaxAttempts        = 10000
)

func ValidatePage(limit, offset int) error {
	if limit < 0 || offset < 0 {
		return fmt.Errorf("clockidcodec: limit and offset must be non-negative")
	}
	if limit > MaxBatchSize {
		return fmt.Errorf("clockidcodec: limit %d exceeds max batch size %d", limit, MaxBatchSize)
	}
	return nil
}

func NormalizePage(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = DefaultPageSize
	}
	if limit > MaxBatchSize {
		limit = MaxBatchSize
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func ValidateID(prefix, id string) error {
	if strings.TrimSpace(prefix) == "" || strings.TrimSpace(id) == "" {
		return fmt.Errorf("clockidcodec: prefix and id are required")
	}
	return nil
}

func ValidateDuration(d, min time.Duration) error {
	if d <= 0 {
		return fmt.Errorf("clockidcodec: duration must be positive")
	}
	if min > 0 && d < min {
		return fmt.Errorf("clockidcodec: duration %s is below minimum %s", d, min)
	}
	return nil
}
