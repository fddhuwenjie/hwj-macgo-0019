package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Chain struct {
	mu   sync.Mutex
	path string
	last Event
}

func OpenChain(path string) (*Chain, error) {
	c := &Chain{path: path}
	if err := c.load(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Chain) load() error {
	if _, err := os.Stat(c.path); os.IsNotExist(err) {
		return nil
	}
	f, err := os.Open(c.path)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	for dec.More() {
		var e Event
		if err := dec.Decode(&e); err != nil {
			return err
		}
		e.Hash = e.ComputeHash()
		c.last = e
	}
	if err := VerifyChainFile(c.path); err != nil {
		return err
	}
	return nil
}

func (c *Chain) Append(ctx context.Context, action, entityType, entityID, actor, rejectReason string, details map[string]string) (*Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	e := Event{
		Index:        c.last.Index + 1,
		Timestamp:    time.Now().UTC(),
		Actor:        actor,
		Action:       action,
		EntityType:   entityType,
		EntityID:     entityID,
		RejectReason: rejectReason,
		Details:      details,
		PrevHash:     c.last.Hash,
	}
	e.Hash = ComputeHash(e.PrevHash, e)

	f, err := os.OpenFile(c.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err := enc.Encode(e); err != nil {
		return nil, err
	}
	if err := f.Sync(); err != nil {
		return nil, err
	}
	c.last = e
	return &e, nil
}

func (c *Chain) LastEvent() Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.last
}

func (c *Chain) Path() string {
	return c.path
}

func (c *Chain) String() string {
	return fmt.Sprintf("audit.Chain(%s)", c.path)
}
