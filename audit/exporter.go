package audit

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

var (
	ErrHashChainBroken = errors.New("audit: hash chain broken")
	ErrIndexGap        = errors.New("audit: index gap")
)

func Export(path string) ([]Event, error) {
	var events []Event
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	for dec.More() {
		var e Event
		if err := dec.Decode(&e); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, Verify(events)
}

func VerifyChainFile(path string) error {
	events, err := Export(path)
	if err != nil {
		return err
	}
	return Verify(events)
}

func Verify(events []Event) error {
	if len(events) == 0 {
		return nil
	}
	if events[0].Index != 1 {
		return fmt.Errorf("%w: first index is %d", ErrIndexGap, events[0].Index)
	}
	if events[0].PrevHash != "" {
		return fmt.Errorf("%w: first prev_hash is not empty", ErrHashChainBroken)
	}
	if events[0].Hash != events[0].ComputeHash() {
		return fmt.Errorf("%w: event %d hash mismatch", ErrHashChainBroken, events[0].Index)
	}
	for i := 1; i < len(events); i++ {
		prev := events[i-1]
		cur := events[i]
		if cur.Index != prev.Index+1 {
			return fmt.Errorf("%w: expected %d got %d", ErrIndexGap, prev.Index+1, cur.Index)
		}
		if cur.PrevHash != prev.Hash {
			return fmt.Errorf("%w: event %d prev_hash mismatch", ErrHashChainBroken, cur.Index)
		}
		if cur.Hash != cur.ComputeHash() {
			return fmt.Errorf("%w: event %d hash mismatch", ErrHashChainBroken, cur.Index)
		}
	}
	return nil
}
