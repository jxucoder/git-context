package storage

import (
	"github.com/jxucoder/git-context/internal/model"
)

// SharedStorage will store data in refs/context/ as git objects so that it can
// be synced with push/pull. Until that exists every method returns
// ErrNotImplemented; the command layer treats that as "nothing stored here".
type SharedStorage struct{}

var _ Storage = (*SharedStorage)(nil)

// NewSharedStorage creates a new shared storage instance.
func NewSharedStorage() *SharedStorage {
	return &SharedStorage{}
}

func (s *SharedStorage) WriteMemory(m *model.Memory) error {
	return ErrNotImplemented
}

func (s *SharedStorage) ReadMemory(id string) (*model.Memory, error) {
	return nil, ErrNotImplemented
}

func (s *SharedStorage) ListMemories() ([]*model.Memory, error) {
	return nil, ErrNotImplemented
}

func (s *SharedStorage) DeleteMemory(id string) error {
	return ErrNotImplemented
}

func (s *SharedStorage) SearchMemories(query string) ([]*model.Memory, error) {
	return nil, ErrNotImplemented
}

func (s *SharedStorage) WriteTask(t *model.Task) error {
	return ErrNotImplemented
}

func (s *SharedStorage) ReadTask(id string) (*model.Task, error) {
	return nil, ErrNotImplemented
}

func (s *SharedStorage) ListTasks() ([]*model.Task, error) {
	return nil, ErrNotImplemented
}

func (s *SharedStorage) UpdateTask(id string, fn func(*model.Task) error) error {
	return ErrNotImplemented
}

func (s *SharedStorage) DeleteTask(id string) error {
	return ErrNotImplemented
}

func (s *SharedStorage) WriteLock(l *model.Lock) error {
	return ErrNotImplemented
}

func (s *SharedStorage) ReadLock(target string) (*model.Lock, error) {
	return nil, ErrNotImplemented
}

func (s *SharedStorage) ListLocks() ([]*model.Lock, error) {
	return nil, ErrNotImplemented
}

func (s *SharedStorage) DeleteLock(target string) error {
	return ErrNotImplemented
}
