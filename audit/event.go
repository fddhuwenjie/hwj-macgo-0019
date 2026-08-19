package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Event struct {
	Index        uint64            `json:"index"`
	Timestamp    time.Time         `json:"timestamp"`
	Actor        string            `json:"actor,omitempty"`
	Action       string            `json:"action"`
	EntityType   string            `json:"entity_type"`
	EntityID     string            `json:"entity_id,omitempty"`
	RejectReason string            `json:"reject_reason,omitempty"`
	Details      map[string]string `json:"details,omitempty"`
	PrevHash     string            `json:"prev_hash"`
	Hash         string            `json:"hash"`
}

func (e Event) canonicalPayload() string {
	keys := make([]string, 0, len(e.Details))
	for k := range e.Details {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d|%d|%s|%s|%s|%s|%s|", e.Index, e.Timestamp.UTC().UnixNano(), e.Actor, e.Action, e.EntityType, e.EntityID, e.RejectReason))
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(e.Details[k])
		sb.WriteString(";")
	}
	return sb.String()
}

func ComputeHash(prevHash string, e Event) string {
	h := sha256.New()
	h.Write([]byte(prevHash))
	h.Write([]byte("|"))
	h.Write([]byte(e.canonicalPayload()))
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

func (e *Event) ComputeHash() string {
	return ComputeHash(e.PrevHash, *e)
}
