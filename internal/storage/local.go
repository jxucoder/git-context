package storage

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jxucoder/git-context/internal/model"
)

// LocalStorage stores data in .git/context/ as plain files.
// This storage is private and never syncs.
type LocalStorage struct {
	baseDir string
}

var _ Storage = (*LocalStorage)(nil)

// NewLocalStorage creates a new local storage instance.
func NewLocalStorage(gitDir string) (*LocalStorage, error) {
	baseDir := filepath.Join(gitDir, "context")

	for _, subdir := range []string{"memory", "tasks", "locks"} {
		dir := filepath.Join(baseDir, subdir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create %s: %w", dir, err)
		}
	}

	return &LocalStorage{baseDir: baseDir}, nil
}

// validateID rejects ids that would resolve outside their storage directory,
// such as "", ".", ".." or anything containing a path separator.
func validateID(id string) error {
	if id == "" || id == "." || id == ".." || strings.ContainsAny(id, `/\`) || filepath.Base(id) != id {
		return fmt.Errorf("%w: %q", ErrInvalidID, id)
	}
	return nil
}

// notFoundOr maps a missing file to ErrNotFound and passes other errors through.
func notFoundOr(id string, err error) error {
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w: %s", ErrNotFound, id)
	}
	return err
}

// corrupt wraps a read failure for an entry that exists on disk.
func corrupt(id string, err error) error {
	if errors.Is(err, ErrCorrupt) {
		return err
	}
	return fmt.Errorf("%w %s: %v", ErrCorrupt, id, err)
}

// Memory operations

func (s *LocalStorage) memoryDir(id string) (string, error) {
	if err := validateID(id); err != nil {
		return "", err
	}
	return filepath.Join(s.baseDir, "memory", id), nil
}

// WriteMemory stores the entry as memory/<id>/meta.json plus content.md.
// content.md is written first so that an entry only becomes visible to
// ListMemories once its metadata is complete.
func (s *LocalStorage) WriteMemory(m *model.Memory) error {
	dir, err := s.memoryDir(m.ID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	meta := *m
	meta.Content = ""
	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dir, "content.md"), []byte(m.Content), 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "meta.json"), metaBytes, 0644)
}

func (s *LocalStorage) ReadMemory(id string) (*model.Memory, error) {
	dir, err := s.memoryDir(id)
	if err != nil {
		return nil, err
	}

	metaBytes, err := os.ReadFile(filepath.Join(dir, "meta.json"))
	if err != nil {
		return nil, notFoundOr(id, err)
	}

	var m model.Memory
	if err := json.Unmarshal(metaBytes, &m); err != nil {
		return nil, corrupt(id, err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "content.md"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	m.Content = string(content)

	return &m, nil
}

func (s *LocalStorage) ListMemories() ([]*model.Memory, error) {
	entries, err := os.ReadDir(filepath.Join(s.baseDir, "memory"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var memories []*model.Memory
	var skipped []error
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		m, err := s.ReadMemory(entry.Name())
		if err != nil {
			skipped = append(skipped, corrupt(entry.Name(), err))
			continue
		}
		memories = append(memories, m)
	}

	return memories, errors.Join(skipped...)
}

func (s *LocalStorage) DeleteMemory(id string) error {
	dir, err := s.memoryDir(id)
	if err != nil {
		return err
	}
	if _, err := os.Stat(dir); err != nil {
		return notFoundOr(id, err)
	}
	return os.RemoveAll(dir)
}

func (s *LocalStorage) SearchMemories(query string) ([]*model.Memory, error) {
	memories, err := s.ListMemories()
	if err != nil && !errors.Is(err, ErrCorrupt) {
		return nil, err
	}

	var results []*model.Memory
	for _, m := range memories {
		if m.MatchesSearch(query) {
			results = append(results, m)
		}
	}

	return results, err
}

// Task operations

func (s *LocalStorage) taskPath(id string) (string, error) {
	if err := validateID(id); err != nil {
		return "", err
	}
	return filepath.Join(s.baseDir, "tasks", id+".json"), nil
}

func (s *LocalStorage) WriteTask(t *model.Task) error {
	path, err := s.taskPath(t.ID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (s *LocalStorage) ReadTask(id string) (*model.Task, error) {
	path, err := s.taskPath(id)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, notFoundOr(id, err)
	}

	var t model.Task
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, corrupt(id, err)
	}

	return &t, nil
}

func (s *LocalStorage) ListTasks() ([]*model.Task, error) {
	entries, err := os.ReadDir(filepath.Join(s.baseDir, "tasks"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var tasks []*model.Task
	var skipped []error
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".json")
		t, err := s.ReadTask(id)
		if err != nil {
			skipped = append(skipped, corrupt(id, err))
			continue
		}
		tasks = append(tasks, t)
	}

	return tasks, errors.Join(skipped...)
}

func (s *LocalStorage) UpdateTask(id string, fn func(*model.Task) error) error {
	t, err := s.ReadTask(id)
	if err != nil {
		return err
	}

	if err := fn(t); err != nil {
		return err
	}

	return s.WriteTask(t)
}

func (s *LocalStorage) DeleteTask(id string) error {
	path, err := s.taskPath(id)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return notFoundOr(id, err)
	}
	return nil
}

// Lock operations

func (s *LocalStorage) lockPath(target string) string {
	return filepath.Join(s.baseDir, "locks", hashTarget(target)+".json")
}

// takeoverGuardTimeout is how long a takeover guard left behind by a crashed
// agent is respected before another agent clears it. Guards are held for the
// few file operations of a stale-lock takeover, so this is generous.
const takeoverGuardTimeout = 30 * time.Second

// WriteLock acquires the lock atomically, so two concurrent callers cannot
// both succeed. An unexpired lock held by someone else yields *LockedError.
// An expired, corrupt, or self-owned lock file is replaced.
func (s *LocalStorage) WriteLock(l *model.Lock) error {
	path := s.lockPath(l.Target)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}

	for attempt := 0; attempt < 100; attempt++ {
		err := createExclusive(path, data)
		if err == nil {
			return nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return err
		}

		existing, err := s.ReadLock(l.Target)
		switch {
		case err == nil && !existing.IsExpired() && !existing.IsOwnedBy(l.LockedBy):
			return &LockedError{Lock: existing}
		case errors.Is(err, ErrNotFound):
			// Removed between our create and read; try to create again.
			continue
		case err != nil && !errors.Is(err, ErrCorrupt):
			return err
		}

		// The file is expired, corrupt, or ours: remove it and create again.
		removed, err := s.removeStaleLock(path, l.LockedBy)
		if err != nil {
			return err
		}
		if !removed {
			// Another agent is in the middle of a takeover; let it finish.
			time.Sleep(time.Duration(1+attempt%5) * time.Millisecond)
		}
	}

	return fmt.Errorf("failed to acquire lock on %s: too many concurrent attempts", l.Target)
}

// createExclusive creates path with data, failing with fs.ErrExist when it
// already exists. The data is written to a temporary file that is then hard
// linked into place, so a concurrent reader never sees a half-written lock.
// Filesystems without hard links fall back to an exclusive create.
func createExclusive(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpPath, 0644); err != nil {
		return err
	}

	err = os.Link(tmpPath, path)
	if err == nil || errors.Is(err, fs.ErrExist) {
		return err
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(path)
		return err
	}
	return f.Close()
}

// removeStaleLock removes the lock file at path if it is still expired,
// corrupt or owned by owner. Takeovers are serialized through a guard file:
// while the guard is held nobody else can remove the file, and nobody can
// create a fresh lock while the stale file exists, so the re-check under the
// guard cannot be invalidated before the removal. It reports false when
// another agent holds the guard.
func (s *LocalStorage) removeStaleLock(path, owner string) (bool, error) {
	guard := path + ".takeover"
	f, err := os.OpenFile(guard, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return false, err
		}
		if info, serr := os.Stat(guard); serr == nil && time.Since(info.ModTime()) > takeoverGuardTimeout {
			os.Remove(guard)
		}
		return false, nil
	}
	f.Close()
	defer os.Remove(guard)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return true, nil
		}
		return false, err
	}
	var l model.Lock
	if json.Unmarshal(data, &l) == nil && !l.IsExpired() && !l.IsOwnedBy(owner) {
		// A fresh lock arrived first; the caller's next create reports it.
		return true, nil
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	return true, nil
}

func (s *LocalStorage) ReadLock(target string) (*model.Lock, error) {
	data, err := os.ReadFile(s.lockPath(target))
	if err != nil {
		return nil, notFoundOr(target, err)
	}

	var l model.Lock
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, corrupt(target, err)
	}

	return &l, nil
}

func (s *LocalStorage) ListLocks() ([]*model.Lock, error) {
	dir := filepath.Join(s.baseDir, "locks")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var locks []*model.Lock
	var skipped []error
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			skipped = append(skipped, corrupt(entry.Name(), err))
			continue
		}
		var l model.Lock
		if err := json.Unmarshal(data, &l); err != nil {
			skipped = append(skipped, corrupt(entry.Name(), err))
			continue
		}
		locks = append(locks, &l)
	}

	return locks, errors.Join(skipped...)
}

func (s *LocalStorage) DeleteLock(target string) error {
	if err := os.Remove(s.lockPath(target)); err != nil {
		return notFoundOr(target, err)
	}
	return nil
}

// Helpers

func hashTarget(target string) string {
	h := sha256.Sum256([]byte(model.NormalizeTarget(target)))
	return fmt.Sprintf("%x", h[:8])
}
