package repository

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sync"
)

type StoreData struct {
	Version int64                      `json:"version"`
	Records map[string]json.RawMessage `json:"records"`
}

type LocalStore struct {
	mu     sync.RWMutex
	path   string
	data   StoreData
	closed bool
}

func NewLocalStore(path string) (*LocalStore, error) {
	s := &LocalStore{
		path: path,
		data: StoreData{
			Records: make(map[string]json.RawMessage),
		},
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *LocalStore) load() error {
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		return s.persistLocked()
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	if len(raw) == 0 {
		s.data = StoreData{Records: make(map[string]json.RawMessage)}
		return nil
	}
	if err := json.Unmarshal(raw, &s.data); err != nil {
		return err
	}
	if s.data.Records == nil {
		s.data.Records = make(map[string]json.RawMessage)
	}
	return nil
}

func (s *LocalStore) persistLocked() error {
	if s.closed {
		return ErrStoreClosed
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".retry-store-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s.data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), s.path)
}

func (s *LocalStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	if err := s.persistLocked(); err != nil {
		return err
	}
	s.closed = true
	return nil
}

func (s *LocalStore) Save(ctx context.Context, prefix string, aggregate interface{}) error {
	tx, err := s.Begin(ctx)
	if err != nil {
		return err
	}
	if err := tx.Save(ctx, prefix, aggregate); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func (s *LocalStore) FindByID(ctx context.Context, prefix string, id string, dest interface{}) error {
	key := prefix + id
	if tx := localTxFromContext(ctx); tx != nil {
		raw, ok := tx.data.Records[key]
		if !ok {
			return ErrNotFound
		}
		return json.Unmarshal(raw, dest)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return ErrStoreClosed
	}
	raw, ok := s.data.Records[key]
	if !ok {
		return ErrNotFound
	}
	return json.Unmarshal(raw, dest)
}

func (s *LocalStore) FindAll(ctx context.Context, prefix string, dest interface{}) error {
	var data StoreData
	if tx := localTxFromContext(ctx); tx != nil {
		data = tx.data
	} else {
		s.mu.RLock()
		if s.closed {
			s.mu.RUnlock()
			return ErrStoreClosed
		}
		data = s.cloneDataLocked()
		s.mu.RUnlock()
	}

	keys := sortedKeys(data.Records, prefix)
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Slice {
		return ErrInvalidDestination
	}
	slice := rv.Elem()
	elemType := slice.Type().Elem()

	for _, key := range keys {
		raw := data.Records[key]
		var elem reflect.Value
		if elemType.Kind() == reflect.Ptr {
			elem = reflect.New(elemType.Elem())
		} else {
			elem = reflect.New(elemType)
		}
		if err := json.Unmarshal(raw, elem.Interface()); err != nil {
			return err
		}
		if elemType.Kind() == reflect.Ptr {
			slice = reflect.Append(slice, elem)
		} else {
			slice = reflect.Append(slice, elem.Elem())
		}
	}
	rv.Elem().Set(slice)
	return nil
}

func (s *LocalStore) cloneDataLocked() StoreData {
	cp := StoreData{
		Version: s.data.Version,
		Records: make(map[string]json.RawMessage, len(s.data.Records)),
	}
	for k, v := range s.data.Records {
		raw := make([]byte, len(v))
		copy(raw, v)
		cp.Records[k] = raw
	}
	return cp
}

func (s *LocalStore) currentVersionLocked(key string) (int64, bool, error) {
	raw, ok := s.data.Records[key]
	if !ok {
		return 0, false, nil
	}
	var probe struct {
		Version int64 `json:"version"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return 0, true, err
	}
	return probe.Version, true, nil
}

type localTxKey struct{}

func WithLocalTransaction(ctx context.Context, tx *LocalTransaction) context.Context {
	return context.WithValue(ctx, localTxKey{}, tx)
}

func localTxFromContext(ctx context.Context) *LocalTransaction {
	tx, _ := ctx.Value(localTxKey{}).(*LocalTransaction)
	return tx
}
