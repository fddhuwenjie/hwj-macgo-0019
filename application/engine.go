package application

import (
	"context"
	"fmt"
	"time"
)

type Engine struct {
	store Store
	clock Clock
	audit AuditSink
}

func NewEngine(store Store, clock Clock, audit AuditSink) *Engine {
	if clock == nil {
		clock = systemClock{}
	}
	return &Engine{store: store, clock: clock, audit: audit}
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

func (e *Engine) Store() Store     { return e.store }
func (e *Engine) Clock() Clock     { return e.clock }
func (e *Engine) Audit() AuditSink { return e.audit }

func (e *Engine) record(ctx context.Context, action, entity, entityID, detail string) {
	if e.audit == nil {
		return
	}
	_ = e.audit.Record(ctx, AuditEvent{
		At:       e.clock.Now(),
		Action:   action,
		Entity:   entity,
		EntityID: entityID,
		Detail:   detail,
	})
}

func withTx(ctx context.Context, store Store, fn func(tx Tx) error) error {
	if store == nil {
		return fmt.Errorf("application: nil store")
	}
	return store.InTx(ctx, fn)
}
