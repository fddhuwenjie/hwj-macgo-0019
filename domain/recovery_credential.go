package domain

import "time"

type RecoveryCredential struct {
	BaseAggregate
	categoryID string
	attemptID  string
	token      string
	issuedAt   time.Time
	expiresAt  time.Time
	redeemedAt *time.Time
}

func NewRecoveryCredential(id, categoryID, attemptID, token string, issuedAt, expiresAt time.Time, now time.Time) (*RecoveryCredential, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	if categoryID == "" || attemptID == "" || token == "" {
		return nil, NewError(ErrorKindInvalidArgument, "NewRecoveryCredential", ErrInvalidInvariant)
	}
	if !issuedAt.Before(expiresAt) {
		return nil, NewError(ErrorKindInvalidArgument, "NewRecoveryCredential", ErrInvalidInvariant)
	}
	if expiresAt.Before(now) {
		return nil, NewError(ErrorKindExpired, "NewRecoveryCredential", ErrRecoveryCredentialExpired)
	}
	c := &RecoveryCredential{
		categoryID: categoryID,
		attemptID:  attemptID,
		token:      token,
		issuedAt:   issuedAt,
		expiresAt:  expiresAt,
	}
	c.setID(id)
	c.setVersion(1)
	c.setCreatedAt(now)
	c.setUpdatedAt(now)
	return c, nil
}

func (c *RecoveryCredential) CategoryID() string {
	if c == nil {
		return ""
	}
	return c.categoryID
}

func (c *RecoveryCredential) AttemptID() string {
	if c == nil {
		return ""
	}
	return c.attemptID
}

func (c *RecoveryCredential) Token() string {
	if c == nil {
		return ""
	}
	return c.token
}

func (c *RecoveryCredential) IssuedAt() time.Time {
	if c == nil {
		return time.Time{}
	}
	return c.issuedAt
}

func (c *RecoveryCredential) ExpiresAt() time.Time {
	if c == nil {
		return time.Time{}
	}
	return c.expiresAt
}

func (c *RecoveryCredential) IsExpired(now time.Time) bool {
	if c == nil {
		return true
	}
	return !now.Before(c.expiresAt)
}

func (c *RecoveryCredential) IsRedeemed() bool {
	if c == nil {
		return true
	}
	return c.redeemedAt != nil
}

func (c *RecoveryCredential) Redeem(now time.Time) error {
	if c == nil {
		return ErrNilArgument
	}
	if c.IsExpired(now) {
		return NewError(ErrorKindExpired, "RecoveryCredential.Redeem", ErrRecoveryCredentialExpired)
	}
	t := now
	c.redeemedAt = &t
	c.bumpVersion(now)
	return nil
}

func (c *RecoveryCredential) Clone() *RecoveryCredential {
	if c == nil {
		return nil
	}
	cp := *c
	cp.redeemedAt = cloneTimePointer(c.redeemedAt)
	return &cp
}
