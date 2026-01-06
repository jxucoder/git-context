package storage

import (
	"github.com/user/git-context/internal/model"
)

// SharedStorage stores data in refs/context/ as git objects.
// This storage syncs with push/pull.
type SharedStorage struct {
	repoPath string
}

// NewSharedStorage creates a new shared storage instance.
func NewSharedStorage(gitDir string) (*SharedStorage, error) {
	// For now, shared storage uses the same file-based approach
	// but stores in a different location. Full git refs implementation
	// will come in Phase 3.
	return &SharedStorage{repoPath: gitDir}, nil
}

// Memory operations - delegate to local for now, will use git refs later

func (s *SharedStorage) WriteMemory(m *model.Memory) error {
	// TODO: Implement git refs storage
	return nil
}

func (s *SharedStorage) ReadMemory(id string) (*model.Memory, error) {
	// TODO: Implement git refs storage
	return nil, nil
}

func (s *SharedStorage) ListMemories() ([]*model.Memory, error) {
	// TODO: Implement git refs storage
	return nil, nil
}

func (s *SharedStorage) DeleteMemory(id string) error {
	// TODO: Implement git refs storage
	return nil
}

func (s *SharedStorage) SearchMemories(query string) ([]*model.Memory, error) {
	// TODO: Implement git refs storage
	return nil, nil
}

// Task operations

func (s *SharedStorage) WriteTask(t *model.Task) error {
	// TODO: Implement git refs storage
	return nil
}

func (s *SharedStorage) ReadTask(id string) (*model.Task, error) {
	// TODO: Implement git refs storage
	return nil, nil
}

func (s *SharedStorage) ListTasks() ([]*model.Task, error) {
	// TODO: Implement git refs storage
	return nil, nil
}

func (s *SharedStorage) UpdateTask(id string, fn func(*model.Task) error) error {
	// TODO: Implement git refs storage
	return nil
}

func (s *SharedStorage) DeleteTask(id string) error {
	// TODO: Implement git refs storage
	return nil
}

// Lock operations

func (s *SharedStorage) WriteLock(l *model.Lock) error {
	// TODO: Implement git refs storage
	return nil
}

func (s *SharedStorage) ReadLock(target string) (*model.Lock, error) {
	// TODO: Implement git refs storage
	return nil, nil
}

func (s *SharedStorage) ListLocks() ([]*model.Lock, error) {
	// TODO: Implement git refs storage
	return nil, nil
}

func (s *SharedStorage) DeleteLock(target string) error {
	// TODO: Implement git refs storage
	return nil
}

