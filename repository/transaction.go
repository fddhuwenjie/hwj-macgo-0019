package repository

import (
	"context"
	"encoding/json"
)

type LocalTransaction struct {
	store     *LocalStore
	data      StoreData
	originals map[string]int64
	dirty     bool
}

func (s *LocalStore) Begin(ctx context.Context) (*LocalTransaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return nil, ErrStoreClosed
	}
	return &LocalTransaction{
		store:     s,
		data:      s.cloneDataLocked(),
		originals: make(map[string]int64),
	}, nil
}

func (tx *LocalTransaction) Save(ctx context.Context, prefix string, aggregate interface{}) error {
	if aggregate == nil {
		return ErrInvalidAggregate
	}
	id := idOf(aggregate)
	if id == "" {
		return ErrInvalidAggregate
	}
	key := prefix + id
	version := versionOf(aggregate)

	if origVer, exists := tx.originals[key]; exists {
		if version != origVer && version != origVer+1 {
			return ErrVersionConflict
		}
		if version == origVer {
			// allow saving same version once; after save we do not update originals,
			// so repeated calls with same original version would conflict on second
			// attempt unless aggregate version was advanced by caller.
			tx.originals[key] = origVer
		} else {
			// already saved once in this transaction; allow continuation.
		}
	} else {
		tx.originals[key] = version
	}

	cp, err := cloneAggregate(aggregate)
	if err != nil {
		return err
	}
	newVersion := version + 1
	setVersion(cp, newVersion)
	raw, err := json.Marshal(cp)
	if err != nil {
		return err
	}
	if tx.data.Records == nil {
		tx.data.Records = make(map[string]json.RawMessage)
	}
	tx.data.Records[key] = raw
	tx.dirty = true
	return nil
}

func (tx *LocalTransaction) Commit(ctx context.Context) error {
	if tx.store == nil {
		return ErrTransactionAborted
	}
	tx.store.mu.Lock()
	defer tx.store.mu.Unlock()
	if tx.store.closed {
		return ErrStoreClosed
	}
	for key, origVer := range tx.originals {
		cur, exists, err := tx.store.currentVersionLocked(key)
		if err != nil {
			return err
		}
		switch {
		case origVer == 0 && exists:
			return ErrDuplicateKey
		case origVer > 0:
			if !exists {
				return ErrVersionConflict
			}
			if cur != origVer && cur != origVer+1 {
				return ErrVersionConflict
			}
		}
	}
	tx.store.data = tx.data
	tx.store.data.Version++
	if err := tx.store.persistLocked(); err != nil {
		return err
	}
	tx.dirty = false
	return nil
}

func (tx *LocalTransaction) Rollback(ctx context.Context) error {
	if tx.store == nil {
		return ErrTransactionAborted
	}
	tx.store.mu.Lock()
	tx.store.data = tx.data
	tx.store.mu.Unlock()
	tx.data = StoreData{Records: make(map[string]json.RawMessage)}
	tx.originals = make(map[string]int64)
	tx.dirty = false
	return nil
}

func (tx *LocalTransaction) Data() StoreData {
	return tx.data
}
