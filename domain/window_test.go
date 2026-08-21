package domain

import (
	"errors"
	"testing"
	"time"
)

func newTestWindow(t *testing.T, limit int) *Window {
	t.Helper()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	w, err := NewWindow("win-1", "pol-1", start, 24*time.Hour, limit, start)
	if err != nil {
		t.Fatalf("NewWindow: %v", err)
	}
	return w
}

func TestCommitReservationRejectsWhenNotReserved(t *testing.T) {
	now := time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)
	w := newTestWindow(t, 10)

	// Window has zero reserved capacity. Committing any amount must be refused
	// rather than driving reserved negative or inflating used.
	err := w.CommitReservation(1, now)
	if err == nil {
		t.Fatal("expected error when committing without reserved capacity")
	}
	if !errors.Is(err, ErrReservationNotFound) {
		t.Fatalf("expected ErrReservationNotFound, got %v", err)
	}
	if w.Reserved() != 0 {
		t.Fatalf("reserved changed to %d; want 0", w.Reserved())
	}
	if w.Used() != 0 {
		t.Fatalf("used changed to %d; want 0", w.Used())
	}
}

func TestCommitReservationRejectsOvercommit(t *testing.T) {
	now := time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)
	w := newTestWindow(t, 10)

	if err := w.Reserve(3, now); err != nil {
		t.Fatalf("Reserve: %v", err)
	}

	// Committing more than the 3 reserved must be refused, not silently
	// decrement reserved below zero or invent used capacity.
	err := w.CommitReservation(4, now)
	if err == nil {
		t.Fatal("expected error when over-committing")
	}
	if !errors.Is(err, ErrReservationNotFound) {
		t.Fatalf("expected ErrReservationNotFound, got %v", err)
	}
	if w.Reserved() != 3 {
		t.Fatalf("reserved changed to %d; want 3", w.Reserved())
	}
	if w.Used() != 0 {
		t.Fatalf("used changed to %d; want 0", w.Used())
	}
}

func TestCommitReservationLegalPathTransfersReservedToUsed(t *testing.T) {
	now := time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)
	w := newTestWindow(t, 10)

	if err := w.Reserve(3, now); err != nil {
		t.Fatalf("Reserve: %v", err)
	}

	// Legal path: commit a portion of what is reserved.
	if err := w.CommitReservation(2, now); err != nil {
		t.Fatalf("CommitReservation: %v", err)
	}
	if w.Reserved() != 1 {
		t.Fatalf("reserved = %d; want 1", w.Reserved())
	}
	if w.Used() != 2 {
		t.Fatalf("used = %d; want 2", w.Used())
	}
	if w.Remaining() != 7 {
		t.Fatalf("remaining = %d; want 7", w.Remaining())
	}

	// Commit the rest; reserved should reach zero, not negative.
	if err := w.CommitReservation(1, now); err != nil {
		t.Fatalf("CommitReservation remainder: %v", err)
	}
	if w.Reserved() != 0 {
		t.Fatalf("reserved = %d; want 0", w.Reserved())
	}
	if w.Used() != 3 {
		t.Fatalf("used = %d; want 3", w.Used())
	}
}

func TestCommitReservationRejectsNonPositive(t *testing.T) {
	now := time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)
	w := newTestWindow(t, 10)

	if err := w.Reserve(2, now); err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if err := w.CommitReservation(0, now); err == nil {
		t.Fatal("expected error for zero commit")
	}
	if err := w.CommitReservation(-1, now); err == nil {
		t.Fatal("expected error for negative commit")
	}
	if w.Reserved() != 2 || w.Used() != 0 {
		t.Fatalf("state mutated by rejected commit: reserved=%d used=%d", w.Reserved(), w.Used())
	}
}
