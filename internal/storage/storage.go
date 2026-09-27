// Package storage provides interfaces and implementations for storing context data.
package storage

import (
	"errors"
	"fmt"

	"github.com/jxucoder/git-context/internal/model"
)

var (
	// ErrNotFound is returned when an entry, task or lock does not exist.
	ErrNotFound = errors.New("not found")
	// ErrNotImplemented is returned by backends that do not exist yet.
	ErrNotImplemented = errors.New("shared storage is not implemented yet")
	// ErrInvalidID is returned for ids that could resolve outside the storage directory.
	ErrInvalidID = errors.New("invalid id")
	// ErrCorrupt wraps entries that exist on disk but cannot be read.
	ErrCorrupt = errors.New("corrupt entry")
)

// LockedError is returned by WriteLock when the target is held by someone else.
type LockedError struct {
	Lock *model.Lock
}

func (e *LockedError) Error() string {
	return fmt.Sprintf("already locked by %s", e.Lock.LockedBy)
}

// Storage defines the interface for storing and retrieving context data.
//
// List and Search methods return every readable item. Entries that exist but
// cannot be read do not fail the call: they are reported through an error
// wrapping ErrCorrupt that is returned alongside the items.
type Storage interface {
	// Memory operations
	WriteMemory(m *model.Memory) error
	ReadMemory(id string) (*model.Memory, error)
	ListMemories() ([]*model.Memory, error)
	DeleteMemory(id string) error
	SearchMemories(query string) ([]*model.Memory, error)

	// Task operations
	WriteTask(t *model.Task) error
	ReadTask(id string) (*model.Task, error)
	ListTasks() ([]*model.Task, error)
	UpdateTask(id string, fn func(*model.Task) error) error
	DeleteTask(id string) error

	// Lock operations
	WriteLock(l *model.Lock) error
	ReadLock(target string) (*model.Lock, error)
	ListLocks() ([]*model.Lock, error)
	DeleteLock(target string) error
}

// MultiStorage combines local and shared storage.
type MultiStorage struct {
	Local  *LocalStorage
	Shared *SharedStorage
}

// NewMultiStorage creates both local and shared storage.
func NewMultiStorage(gitDir string) (*MultiStorage, error) {
	local, err := NewLocalStorage(gitDir)
	if err != nil {
		return nil, err
	}

	return &MultiStorage{
		Local:  local,
		Shared: NewSharedStorage(),
	}, nil
}
