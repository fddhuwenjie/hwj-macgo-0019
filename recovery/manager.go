package recovery

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

// Manager coordinates snapshot loading and journal replay.
type Manager struct {
	snapshots Snapshotter
	logPath   string
	store     RecoveryStore
	now       func() time.Time
}

func NewManager(snapshotDir string, logPath string, store RecoveryStore) *Manager {
	return &Manager{
		snapshots: NewFileSnapshotter(snapshotDir),
		logPath:   logPath,
		store:     store,
		now:       time.Now,
	}
}

func NewRecoveryManager(snapshotDir string, logPath string, store RecoveryStore) *Manager {
	return NewManager(snapshotDir, logPath, store)
}

// Recover loads the latest snapshot and applies journal entries after it.
func (m *Manager) Recover(ctx context.Context) (Snapshot, error) {
	names, err := m.snapshots.ListSnapshots(ctx)
	if err != nil {
		return Snapshot{}, fmt.Errorf("recovery: list snapshots: %w", err)
	}
	var snap Snapshot
	if len(names) > 0 {
		name := names[len(names)-1]
		snap, err = m.snapshots.ReadSnapshot(ctx, name)
		if err != nil {
			return Snapshot{}, fmt.Errorf("recovery: read latest snapshot %s: %w", name, err)
		}
	}

	reader, err := OpenFileJournalReader(m.logPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return snap, nil
		}
		return Snapshot{}, fmt.Errorf("recovery: open journal: %w", err)
	}
	defer reader.Close()

	replayer := NewReplayer(m.store)
	result, err := replayer.Apply(ctx, reader)
	if err != nil {
		return Snapshot{}, fmt.Errorf("recovery: replay: %w", err)
	}
	if result.LastVersion > snap.Version {
		snap.Version = result.LastVersion
	}
	snap.JournalSequence = result.LastSequence
	return snap, nil
}

// Snapshot writes the current recovered state through the snapshotter.
func (m *Manager) Snapshot(ctx context.Context, snap Snapshot) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return m.snapshots.WriteSnapshot(ctx, snap)
}

// Verify checks the snapshot consistency.
func (m *Manager) Verify(snap Snapshot) error {
	return VerifySnapshot(snap)
}
