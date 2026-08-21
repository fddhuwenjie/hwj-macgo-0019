package domain

import (
	"errors"
	"testing"
	"time"
)

func newTestCredential(t *testing.T, issuedAt, expiresAt, now time.Time) *RecoveryCredential {
	t.Helper()
	c, err := NewRecoveryCredential("cred-1", "cat-1", "att-1", "token-1", issuedAt, expiresAt, now)
	if err != nil {
		t.Fatalf("NewRecoveryCredential: %v", err)
	}
	return c
}

// A credential must be redeemable exactly once. A second redeem of the same
// credential must be rejected as an invalid state transition and must not
// advance the aggregate version.
func TestRecoveryCredentialRedeemIsOnce(t *testing.T) {
	issued := time.Unix(1000, 0)
	expires := time.Unix(2000, 0)
	c := newTestCredential(t, issued, expires, issued)

	if c.IsRedeemed() {
		t.Fatal("fresh credential must not be redeemed")
	}
	v0 := c.Version()

	if err := c.Redeem(issued.Add(10)); err != nil {
		t.Fatalf("first redeem: %v", err)
	}
	if !c.IsRedeemed() {
		t.Fatal("credential not marked redeemed after first redeem")
	}
	v1 := c.Version()
	if v1 != v0+1 {
		t.Fatalf("first redeem should bump version once: got %d -> %d", v0, v1)
	}

	err := c.Redeem(issued.Add(20))
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("second redeem must fail with ErrInvalidTransition, got %v", err)
	}
	v2 := c.Version()
	if v2 != v1 {
		t.Fatalf("failed second redeem must not bump version: %d -> %d", v1, v2)
	}
}

// A credential at its exact expiry instant must be treated as expired, so
// redeeming it fails with the expired error and does not change state.
func TestRecoveryCredentialExpiredAtExpiryInstant(t *testing.T) {
	issued := time.Unix(1000, 0)
	expires := time.Unix(2000, 0)
	c := newTestCredential(t, issued, expires, issued)

	if !c.IsExpired(expires) {
		t.Fatalf("IsExpired(expiresAt) must be true, got false")
	}
	v0 := c.Version()

	err := c.Redeem(expires)
	if !errors.Is(err, ErrRecoveryCredentialExpired) {
		t.Fatalf("redeem at expiry must fail with ErrRecoveryCredentialExpired, got %v", err)
	}
	if c.IsRedeemed() {
		t.Fatal("expired redeem must not mark credential redeemed")
	}
	if c.Version() != v0 {
		t.Fatalf("expired redeem must not bump version: %d -> %d", v0, c.Version())
	}

	// just before expiry is still valid and redeemable once
	if err := c.Redeem(expires.Add(-1)); err != nil {
		t.Fatalf("redeem just before expiry must succeed, got %v", err)
	}
}

// ValidateRecoveryCredential must reject an already-redeemed credential.
func TestValidateRecoveryCredentialRejectsRedeemed(t *testing.T) {
	issued := time.Unix(1000, 0)
	expires := time.Unix(2000, 0)
	now := issued.Add(5)
	c := newTestCredential(t, issued, expires, issued)

	if err := ValidateRecoveryCredential(c, "cat-1", now); err != nil {
		t.Fatalf("precondition check on fresh credential: %v", err)
	}

	if err := c.Redeem(now); err != nil {
		t.Fatalf("redeem: %v", err)
	}

	err := ValidateRecoveryCredential(c, "cat-1", now.Add(1))
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("validate after redeem must fail with ErrInvalidTransition, got %v", err)
	}
}
