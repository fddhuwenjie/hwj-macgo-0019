package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// offlineSelfCheck runs a minimal offline verification that does not require
// external services. It is invoked by --self-check.
func offlineSelfCheck(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	var failed bool
	fail := func(format string, args ...any) {
		failed = true
		msg := fmt.Sprintf("self-check: "+format, args...)
		fmt.Fprintln(os.Stderr, msg)
	}

	if runtime.GOOS == "" {
		fail("runtime.GOOS is empty")
	}
	tmp, err := os.MkdirTemp("", "hwj-macgo-0019-selfcheck-*")
	if err != nil {
		fail("create temp dir: %v", err)
	} else {
		defer os.RemoveAll(tmp)
		p := filepath.Join(tmp, "probe")
		if err := os.WriteFile(p, []byte("ok"), 0o600); err != nil {
			fail("write probe: %v", err)
		}
		if b, err := os.ReadFile(p); err != nil || string(b) != "ok" {
			fail("read probe: got %q err=%v", b, err)
		}
	}
	select {
	case <-ctx.Done():
		fail("context canceled during self-check: %v", ctx.Err())
	case <-time.After(0):
	}

	if failed {
		return errors.New("self-check completed with failures")
	}
	fmt.Fprintln(os.Stdout, "self-check: OK")
	return nil
}
