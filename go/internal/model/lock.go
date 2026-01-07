package model

import "time"

// DefaultLockExpiry is how long locks last before expiring.
const DefaultLockExpiry = 4 * time.Hour

// Lock represents a lock on a task or file path.
type Lock struct {
	Target    string    `json:"target"`
	LockedBy  string    `json:"lockedBy"`
	LockedAt  time.Time `json:"lockedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// NewLock creates a new lock with default expiry.
func NewLock(target, lockedBy string) *Lock {
	now := time.Now().UTC()
	return &Lock{
		Target:    target,
		LockedBy:  lockedBy,
		LockedAt:  now,
		ExpiresAt: now.Add(DefaultLockExpiry),
	}
}

// IsExpired returns true if the lock has expired.
func (l *Lock) IsExpired() bool {
	return time.Now().UTC().After(l.ExpiresAt)
}

// IsOwnedBy returns true if the lock is owned by the given user.
func (l *Lock) IsOwnedBy(user string) bool {
	return l.LockedBy == user
}


