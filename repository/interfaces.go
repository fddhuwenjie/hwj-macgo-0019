package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var (
	ErrInvalid          = errors.New(`repository: invalid argument`)
	ErrDocumentNotFound = os.ErrNotExist
	ErrConflict         = os.ErrExist
	ErrClosed           = errors.New(`repository: store closed`)
	ErrCorrupt          = errors.New(`repository: corrupted store data`)
)

type Store interface {
	Get(ctx context.Context, id string) (json.RawMessage, int64, error)
	Put(ctx context.Context, id string, expectedVersion int64, data json.RawMessage) (int64, error)
	Delete(ctx context.Context, id string, expectedVersion int64) error
	List(ctx context.Context, prefix string) ([]string, error)
	Close() error
}

type listEntry struct {
	ID      string
	Version int64
	Data    json.RawMessage
}

type fileEnvelope struct {
	ID      string
	Version int64
	Data    json.RawMessage
}

type MemoryStore struct {
	mu       sync.Mutex
	docs     map[string]json.RawMessage
	versions map[string]int64
	closed   bool
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{docs: map[string]json.RawMessage{}, versions: map[string]int64{}}
}

func (m *MemoryStore) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *MemoryStore) Get(ctx context.Context, id string) (json.RawMessage, int64, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, 0, err
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, 0, ErrClosed
	}
	data, ok := m.docs[id]
	if !ok {
		return nil, 0, os.ErrNotExist
	}
	out := make(json.RawMessage, len(data))
	copy(out, data)
	return out, m.versions[id], nil
}

func (m *MemoryStore) Put(ctx context.Context, id string, expectedVersion int64, data json.RawMessage) (int64, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, ErrClosed
	}
	version, ok := m.versions[id]
	if !ok {
		version = 0
	}
	if version != expectedVersion {
		return 0, os.ErrExist
	}
	out := make(json.RawMessage, len(data))
	copy(out, data)
	m.docs[id] = out
	m.versions[id] = version + 1
	return version + 1, nil
}

func (m *MemoryStore) Delete(ctx context.Context, id string, expectedVersion int64) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return ErrClosed
	}
	version, ok := m.versions[id]
	if !ok {
		return nil
	}
	if version != expectedVersion {
		return os.ErrExist
	}
	delete(m.docs, id)
	delete(m.versions, id)
	return nil
}

func (m *MemoryStore) List(ctx context.Context, prefix string) ([]string, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, ErrClosed
	}
	out := make([]string, 0)
	for id, data := range m.docs {
		if !strings.HasPrefix(id, prefix) {
			continue
		}
		le := listEntry{ID: id, Version: m.versions[id], Data: data}
		b, err := json.Marshal(le)
		if err != nil {
			return nil, err
		}
		out = append(out, string(b))
	}
	sort.Strings(out)
	return out, nil
}

type FileStore struct {
	path   string
	mu     sync.Mutex
	closed bool
}

func NewFileStore(path string) (*FileStore, error) {
	if path == `` {
		return nil, ErrInvalid
	}
	if err := os.MkdirAll(filepath.Join(path, `docs`), 0755); err != nil {
		return nil, err
	}
	return &FileStore{path: path}, nil
}

func OpenFileStore(path string) (*FileStore, error) {
	return NewFileStore(path)
}

func New(path string) (*FileStore, error) {
	return NewFileStore(path)
}

func (f *FileStore) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

func (f *FileStore) docPath(id string) string {
	sum := sha256.Sum256([]byte(id))
	return filepath.Join(f.path, `docs`, hex.EncodeToString(sum[:])+`.doc`)
}

func (f *FileStore) Get(ctx context.Context, id string) (json.RawMessage, int64, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, 0, err
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return nil, 0, ErrClosed
	}
	b, err := os.ReadFile(f.docPath(id))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, os.ErrNotExist
		}
		return nil, 0, err
	}
	var env fileEnvelope
	if err := json.Unmarshal(b, &env); err != nil {
		return nil, 0, ErrCorrupt
	}
	if env.ID != id {
		return nil, 0, ErrCorrupt
	}
	out := make(json.RawMessage, len(env.Data))
	copy(out, env.Data)
	return out, env.Version, nil
}

func (f *FileStore) Put(ctx context.Context, id string, expectedVersion int64, data json.RawMessage) (int64, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, ErrClosed
	}
	path := f.docPath(id)
	if b, err := os.ReadFile(path); err == nil {
		var env fileEnvelope
		if err := json.Unmarshal(b, &env); err != nil {
			return 0, ErrCorrupt
		}
	} else if !os.IsNotExist(err) {
		return 0, err
	}
	env := fileEnvelope{ID: id, Version: expectedVersion + 1, Data: data}
	b, err := json.Marshal(env)
	if err != nil {
		return 0, err
	}
	tmp := path + `.tmp`
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return 0, err
	}
	if err := os.Rename(tmp, path); err != nil {
		if rmErr := os.Remove(tmp); rmErr != nil && !os.IsNotExist(rmErr) {
			return 0, rmErr
		}
		return 0, err
	}
	return env.Version, nil
}

func (f *FileStore) Delete(ctx context.Context, id string, expectedVersion int64) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return ErrClosed
	}
	path := f.docPath(id)
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var env fileEnvelope
	if err := json.Unmarshal(b, &env); err != nil {
		return ErrCorrupt
	}
	if env.Version != expectedVersion {
		return os.ErrExist
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (f *FileStore) List(ctx context.Context, prefix string) ([]string, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return nil, ErrClosed
	}
	dir := filepath.Join(f.path, `docs`)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]string, 0)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), `.doc`) {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		var env fileEnvelope
		if err := json.Unmarshal(b, &env); err != nil {
			return nil, ErrCorrupt
		}
		if env.ID == `` || !strings.HasPrefix(env.ID, prefix) {
			continue
		}
		data := make(json.RawMessage, len(env.Data))
		copy(data, env.Data)
		le := listEntry{ID: env.ID, Version: env.Version, Data: data}
		b, err = json.Marshal(le)
		if err != nil {
			return nil, err
		}
		out = append(out, string(b))
	}
	sort.Strings(out)
	return out, nil
}
