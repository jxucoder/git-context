// Package storage provides interfaces and implementations for storing context data.
package storage

import "github.com/user/git-context/internal/model"

// Storage defines the interface for storing and retrieving context data.
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
	Local  Storage
	Shared Storage
}

// NewMultiStorage creates a new multi-storage instance.
func NewMultiStorage(gitDir string) (*MultiStorage, error) {
	local, err := NewLocalStorage(gitDir)
	if err != nil {
		return nil, err
	}
	
	shared, err := NewSharedStorage(gitDir)
	if err != nil {
		return nil, err
	}
	
	return &MultiStorage{
		Local:  local,
		Shared: shared,
	}, nil
}


