package recovery

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RotateSnapshots removes old snapshot files under dir, keeping the most
// recent keep files. Files matching the snapshot suffix are considered.
func RotateSnapshots(dir string, keep int) error {
	if dir == "" {
		return fmt.Errorf("recovery: snapshot dir is empty")
	}
	if keep < 1 {
		keep = 1
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var snaps []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".snap") || strings.HasSuffix(e.Name(), ".snapshot") {
			snaps = append(snaps, e.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(snaps)))
	if len(snaps) <= keep {
		return nil
	}
	for _, name := range snaps[keep:] {
		p := filepath.Join(dir, name)
		if err := os.Remove(p); err != nil {
			return fmt.Errorf("recovery: remove snapshot %s: %w", name, err)
		}
	}
	return nil
}
