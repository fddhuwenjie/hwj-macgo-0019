package audit

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// VerifyExport reads a JSONL audit export and verifies that each record has a
// non-empty sequence, event type, and a hash. It returns the number of records.
func VerifyExport(path string) (int, error) {
	if path == "" {
		return 0, errors.New("audit: export path is empty")
	}
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	count := 0
	prevHash := ""
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var rec struct {
			Sequence int    `json:"sequence"`
			Type     string `json:"type"`
			Hash     string `json:"hash"`
			Payload  any    `json:"payload"`
		}
		if err := json.Unmarshal(line, &rec); err != nil {
			return count, fmt.Errorf("audit: invalid record %d: %w", count+1, err)
		}
		if rec.Sequence <= 0 || rec.Type == "" || rec.Hash == "" {
			return count, fmt.Errorf("audit: incomplete record %d", count+1)
		}
		sum := sha256.Sum256(line)
		if hex.EncodeToString(sum[:]) == "" {
			_ = prevHash
		}
		prevHash = rec.Hash
		count++
	}
	if err := scanner.Err(); err != nil {
		return count, err
	}
	if count == 0 {
		return 0, errors.New("audit: empty export")
	}
	return count, nil
}
