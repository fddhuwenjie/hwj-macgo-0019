package clockidcodec

import (
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type IDGenerator interface {
	NewID(prefix string) string
}

type RandomIDGenerator struct {
	mu sync.Mutex
}

func NewRandomIDGenerator() *RandomIDGenerator {
	return &RandomIDGenerator{}
}

func (g *RandomIDGenerator) NewID(prefix string) string {
	g.mu.Lock()
	defer g.mu.Unlock()
	var b [16]byte
	if _, err := crand.Read(b[:]); err == nil {
		return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b[:]))
	}
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
