package domain

import (
	"strings"
	"time"
)

type BaseAggregate struct {
	id        string
	version   int64
	createdAt time.Time
	updatedAt time.Time
}

func (b BaseAggregate) ID() string {
	return b.id
}

func (b BaseAggregate) Version() int64 {
	return b.version
}

func (b BaseAggregate) CreatedAt() time.Time {
	return b.createdAt
}

func (b BaseAggregate) UpdatedAt() time.Time {
	return b.updatedAt
}

func (b *BaseAggregate) setID(id string) {
	b.id = id
}

func (b *BaseAggregate) setVersion(version int64) {
	b.version = version
}

func (b *BaseAggregate) setCreatedAt(t time.Time) {
	b.createdAt = t
}

func (b *BaseAggregate) setUpdatedAt(t time.Time) {
	b.updatedAt = t
}

func (b *BaseAggregate) bumpVersion(now time.Time) {
	b.version++
	b.updatedAt = now
}

func validateID(id string) error {
	if strings.TrimSpace(id) == "" {
		return NewError(ErrorKindInvalidArgument, "validateID", ErrInvalidInvariant)
	}
	return nil
}

func validateNotNil(v interface{}) error {
	if v == nil {
		return ErrNilArgument
	}
	return nil
}

func cloneStringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneTimePointer(src *time.Time) *time.Time {
	if src == nil {
		return nil
	}
	t := *src
	return &t
}
