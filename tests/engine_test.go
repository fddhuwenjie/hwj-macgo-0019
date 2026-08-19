package tests

import (
	"context"
	"testing"
	"time"
)

func TestContextCancellationSmoke(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("expected cancellation")
	}
}
