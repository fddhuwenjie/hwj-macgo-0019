package clockidcodec

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

type OffsetClock struct {
	mu     sync.RWMutex
	base   time.Time
	offset time.Duration
}

func NewOffsetClock(base time.Time, offset time.Duration) *OffsetClock {
	return &OffsetClock{base: base, offset: offset}
}

func (c *OffsetClock) Now() time.Time {
	if c == nil {
		return time.Time{}
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.base.Add(time.Since(c.base) + c.offset)
}

func (c *OffsetClock) Advance(d time.Duration) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.offset += d
}

func (c *OffsetClock) SetOffset(d time.Duration) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.offset = d
}
