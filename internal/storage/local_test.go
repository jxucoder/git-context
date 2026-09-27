package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jxucoder/git-context/internal/model"
)

func newTestStorage(t *testing.T) *LocalStorage {
	t.Helper()
	s, err := NewLocalStorage(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocalStorage: %v", err)
	}
	return s
}

func mustWriteMemory(t *testing.T, s *LocalStorage, m *model.Memory) {
	t.Helper()
	if err := s.WriteMemory(m); err != nil {
		t.Fatalf("WriteMemory(%s): %v", m.ID, err)
	}
}

func mustWriteTask(t *testing.T, s *LocalStorage, task *model.Task) {
	t.Helper()
	if err := s.WriteTask(task); err != nil {
		t.Fatalf("WriteTask(%s): %v", task.ID, err)
	}
}

// Memory CRUD

func TestLocalStorage_Memory_RoundTrip(t *testing.T) {
	s := newTestStorage(t)
	m := model.NewMemory("Why JWT", "Because stateless", "alice", false)
	m.Tags = []string{"auth"}
	mustWriteMemory(t, s, m)

	got, err := s.ReadMemory(m.ID)
	if err != nil {
		t.Fatalf("ReadMemory: %v", err)
	}
	if got.ID != m.ID || got.Title != m.Title || got.Content != m.Content || got.Author != m.Author {
		t.Errorf("round trip mismatch: got %+v, want %+v", got, m)
	}
	if len(got.Tags) != 1 || got.Tags[0] != "auth" {
		t.Errorf("Tags: got %v, want [auth]", got.Tags)
	}
	if !got.CreatedAt.Equal(m.CreatedAt) || !got.UpdatedAt.Equal(m.UpdatedAt) {
		t.Errorf("timestamps did not round trip: got %v/%v, want %v/%v", got.CreatedAt, got.UpdatedAt, m.CreatedAt, m.UpdatedAt)
	}
}

func TestLocalStorage_Memory_ReadsLegacyMetadata(t *testing.T) {
	s := newTestStorage(t)
	dir := filepath.Join(s.baseDir, "memory", "legacy01")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	// Format written by v0.2 and the original Bash script: second precision, literal Z.
	meta := `{"id":"legacy01","title":"Old","author":"bob","createdAt":"2026-01-06T03:29:42Z","updatedAt":"2026-01-06T03:29:42Z","shared":false}`
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), []byte(meta), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "content.md"), []byte("body"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := s.ReadMemory("legacy01")
	if err != nil {
		t.Fatalf("ReadMemory: %v", err)
	}
	want := time.Date(2026, 1, 6, 3, 29, 42, 0, time.UTC)
	if !got.CreatedAt.Equal(want) || got.Content != "body" {
		t.Errorf("got createdAt=%v content=%q", got.CreatedAt, got.Content)
	}
}

func TestLocalStorage_Memory_NotFound(t *testing.T) {
	s := newTestStorage(t)

	if _, err := s.ReadMemory("nonexistent"); !errors.Is(err, ErrNotFound) {
		t.Errorf("ReadMemory: got %v, want ErrNotFound", err)
	}
	if err := s.DeleteMemory("nonexistent"); !errors.Is(err, ErrNotFound) {
		t.Errorf("DeleteMemory: got %v, want ErrNotFound", err)
	}
}

func TestLocalStorage_Memory_Delete(t *testing.T) {
	s := newTestStorage(t)
	m := model.NewMemory("To delete", "content", "bob", false)
	mustWriteMemory(t, s, m)

	if err := s.DeleteMemory(m.ID); err != nil {
		t.Fatalf("DeleteMemory: %v", err)
	}
	if _, err := s.ReadMemory(m.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("after delete: got %v, want ErrNotFound", err)
	}
}

func TestLocalStorage_RejectsUnsafeIDs(t *testing.T) {
	s := newTestStorage(t)
	keep := model.NewMemory("keep", "content", "alice", false)
	mustWriteMemory(t, s, keep)

	for _, id := range []string{"", ".", "..", "../..", "../tasks", "a/b", `a\b`, "/abs"} {
		if _, err := s.ReadMemory(id); !errors.Is(err, ErrInvalidID) {
			t.Errorf("ReadMemory(%q): got %v, want ErrInvalidID", id, err)
		}
		if err := s.DeleteMemory(id); !errors.Is(err, ErrInvalidID) {
			t.Errorf("DeleteMemory(%q): got %v, want ErrInvalidID", id, err)
		}
		if err := s.WriteMemory(&model.Memory{ID: id}); !errors.Is(err, ErrInvalidID) {
			t.Errorf("WriteMemory(%q): got %v, want ErrInvalidID", id, err)
		}
		if _, err := s.ReadTask(id); !errors.Is(err, ErrInvalidID) {
			t.Errorf("ReadTask(%q): got %v, want ErrInvalidID", id, err)
		}
		if err := s.WriteTask(&model.Task{ID: id}); !errors.Is(err, ErrInvalidID) {
			t.Errorf("WriteTask(%q): got %v, want ErrInvalidID", id, err)
		}
		if err := s.DeleteTask(id); !errors.Is(err, ErrInvalidID) {
			t.Errorf("DeleteTask(%q): got %v, want ErrInvalidID", id, err)
		}
	}

	// Nothing outside memory/<id> may have been touched.
	for _, sub := range []string{"memory", "tasks", "locks"} {
		if _, err := os.Stat(filepath.Join(s.baseDir, sub)); err != nil {
			t.Errorf("%s directory is gone: %v", sub, err)
		}
	}
	if _, err := s.ReadMemory(keep.ID); err != nil {
		t.Errorf("existing entry was removed: %v", err)
	}
}

func TestLocalStorage_Memory_List(t *testing.T) {
	s := newTestStorage(t)
	for _, title := range []string{"first", "second", "third"} {
		mustWriteMemory(t, s, model.NewMemory(title, "content", "alice", false))
	}

	memories, err := s.ListMemories()
	if err != nil {
		t.Fatalf("ListMemories: %v", err)
	}
	if len(memories) != 3 {
		t.Fatalf("expected 3 memories, got %d", len(memories))
	}
}

func TestLocalStorage_Memory_Search(t *testing.T) {
	s := newTestStorage(t)
	mustWriteMemory(t, s, model.NewMemory("JWT Auth", "stateless tokens", "alice", false))
	mustWriteMemory(t, s, model.NewMemory("Database choice", "PostgreSQL", "alice", false))

	results, err := s.SearchMemories("jwt")
	if err != nil {
		t.Fatalf("SearchMemories: %v", err)
	}
	if len(results) != 1 || results[0].Title != "JWT Auth" {
		t.Fatalf("expected [JWT Auth], got %v", results)
	}
}

func TestLocalStorage_CorruptEntryReported(t *testing.T) {
	s := newTestStorage(t)
	good := model.NewMemory("Good entry", "content", "alice", false)
	mustWriteMemory(t, s, good)

	corruptDir := filepath.Join(s.baseDir, "memory", "corrupt1")
	if err := os.MkdirAll(corruptDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(corruptDir, "meta.json"), []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	memories, err := s.ListMemories()
	if !errors.Is(err, ErrCorrupt) {
		t.Errorf("ListMemories error: got %v, want ErrCorrupt", err)
	}
	if len(memories) != 1 || memories[0].ID != good.ID {
		t.Errorf("expected only the valid entry, got %v", memories)
	}

	results, err := s.SearchMemories("content")
	if !errors.Is(err, ErrCorrupt) {
		t.Errorf("SearchMemories error: got %v, want ErrCorrupt", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 search result, got %d", len(results))
	}

	if _, err := s.ReadMemory("corrupt1"); !errors.Is(err, ErrCorrupt) {
		t.Errorf("ReadMemory(corrupt1): got %v, want ErrCorrupt", err)
	}
}

// Task CRUD

func TestLocalStorage_Task_RoundTrip(t *testing.T) {
	s := newTestStorage(t)
	task := model.NewTask("Implement auth", "OAuth2 flow", "bob", false)
	mustWriteTask(t, s, task)

	got, err := s.ReadTask(task.ID)
	if err != nil {
		t.Fatalf("ReadTask: %v", err)
	}
	if got.Title != task.Title || got.Status != model.TaskOpen || !got.CreatedAt.Equal(task.CreatedAt) {
		t.Errorf("round trip mismatch: got %+v", got)
	}

	if _, err := s.ReadTask("task-missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("ReadTask(missing): got %v, want ErrNotFound", err)
	}
	if err := s.DeleteTask("task-missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("DeleteTask(missing): got %v, want ErrNotFound", err)
	}
}

func TestLocalStorage_Task_Update(t *testing.T) {
	s := newTestStorage(t)
	task := model.NewTask("Fix bug", "", "carol", false)
	mustWriteTask(t, s, task)

	if err := s.UpdateTask(task.ID, func(t *model.Task) error { return t.Claim("carol") }); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	got, err := s.ReadTask(task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != model.TaskClaimed || got.Owner != "carol" {
		t.Errorf("expected claimed by carol, got %s/%s", got.Status, got.Owner)
	}

	// Errors from fn are passed through and nothing is written.
	err = s.UpdateTask(task.ID, func(t *model.Task) error { return t.Claim("dave") })
	var claimed *model.ClaimedError
	if !errors.As(err, &claimed) || claimed.Owner != "carol" {
		t.Errorf("expected ClaimedError{carol}, got %v", err)
	}

	if err := s.UpdateTask("task-missing", func(*model.Task) error { return nil }); !errors.Is(err, ErrNotFound) {
		t.Errorf("UpdateTask(missing): got %v, want ErrNotFound", err)
	}
}

func TestLocalStorage_Task_ListSkipsCorrupt(t *testing.T) {
	s := newTestStorage(t)
	for i := 0; i < 3; i++ {
		mustWriteTask(t, s, model.NewTask("task", "", "alice", false))
	}
	if err := os.WriteFile(filepath.Join(s.baseDir, "tasks", "task-bad.json"), []byte("{"), 0644); err != nil {
		t.Fatal(err)
	}

	tasks, err := s.ListTasks()
	if !errors.Is(err, ErrCorrupt) {
		t.Errorf("ListTasks error: got %v, want ErrCorrupt", err)
	}
	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}
}

// Lock operations

func TestLocalStorage_Lock_RoundTrip(t *testing.T) {
	s := newTestStorage(t)
	if err := s.WriteLock(model.NewLock("src/auth/", "dave")); err != nil {
		t.Fatalf("WriteLock: %v", err)
	}

	// Equivalent spellings resolve to the same lock.
	for _, target := range []string{"src/auth/", "src/auth", "./src/auth"} {
		got, err := s.ReadLock(target)
		if err != nil {
			t.Fatalf("ReadLock(%q): %v", target, err)
		}
		if got.LockedBy != "dave" || got.Target != "src/auth" {
			t.Errorf("ReadLock(%q): got %+v", target, got)
		}
	}

	if _, err := s.ReadLock("other"); !errors.Is(err, ErrNotFound) {
		t.Errorf("ReadLock(other): got %v, want ErrNotFound", err)
	}
	if err := s.DeleteLock("./src/auth"); err != nil {
		t.Fatalf("DeleteLock: %v", err)
	}
	if err := s.DeleteLock("src/auth"); !errors.Is(err, ErrNotFound) {
		t.Errorf("DeleteLock again: got %v, want ErrNotFound", err)
	}
}

func TestLocalStorage_Lock_Exclusive(t *testing.T) {
	s := newTestStorage(t)
	if err := s.WriteLock(model.NewLock("task-abc", "eve")); err != nil {
		t.Fatalf("WriteLock: %v", err)
	}

	err := s.WriteLock(model.NewLock("task-abc", "frank"))
	var locked *LockedError
	if !errors.As(err, &locked) || locked.Lock.LockedBy != "eve" {
		t.Fatalf("expected LockedError held by eve, got %v", err)
	}

	// The holder can refresh their own lock.
	refreshed := model.NewLock("task-abc", "eve")
	refreshed.ExpiresAt = refreshed.ExpiresAt.Add(time.Hour)
	if err := s.WriteLock(refreshed); err != nil {
		t.Fatalf("refresh by owner: %v", err)
	}
	got, err := s.ReadLock("task-abc")
	if err != nil {
		t.Fatal(err)
	}
	if !got.ExpiresAt.Equal(refreshed.ExpiresAt) {
		t.Errorf("refresh did not update expiry: got %v, want %v", got.ExpiresAt, refreshed.ExpiresAt)
	}
}

func TestLocalStorage_Lock_ReplacesExpiredAndCorrupt(t *testing.T) {
	s := newTestStorage(t)

	expired := model.NewLock("src/x", "eve")
	expired.ExpiresAt = time.Now().UTC().Add(-time.Minute)
	if err := s.WriteLock(expired); err != nil {
		t.Fatal(err)
	}
	if err := s.WriteLock(model.NewLock("src/x", "frank")); err != nil {
		t.Errorf("expired lock should be replaced, got %v", err)
	}
	if got, _ := s.ReadLock("src/x"); got == nil || got.LockedBy != "frank" {
		t.Errorf("expected frank to hold src/x, got %+v", got)
	}

	if err := os.WriteFile(s.lockPath("src/y"), []byte("garbage"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ReadLock("src/y"); !errors.Is(err, ErrCorrupt) {
		t.Errorf("ReadLock(corrupt): got %v, want ErrCorrupt", err)
	}
	if err := s.WriteLock(model.NewLock("src/y", "grace")); err != nil {
		t.Errorf("corrupt lock should be replaced, got %v", err)
	}

	locks, err := s.ListLocks()
	if err != nil {
		t.Errorf("ListLocks: %v", err)
	}
	if len(locks) != 2 {
		t.Errorf("expected 2 locks, got %d", len(locks))
	}
}

func TestLocalStorage_Lock_ConcurrentAcquire(t *testing.T) {
	s := newTestStorage(t)
	const agents = 32

	stale := model.NewLock("shared/target", "zombie")
	stale.ExpiresAt = time.Now().UTC().Add(-time.Hour)
	if err := s.WriteLock(stale); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	results := make(chan error, agents)
	for i := 0; i < agents; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results <- s.WriteLock(model.NewLock("shared/target", fmt.Sprintf("agent%02d", i)))
		}(i)
	}
	wg.Wait()
	close(results)

	wins := 0
	for err := range results {
		var locked *LockedError
		switch {
		case err == nil:
			wins++
		case errors.As(err, &locked):
		default:
			t.Errorf("unexpected error: %v", err)
		}
	}
	if wins != 1 {
		t.Errorf("expected exactly one agent to acquire the lock, got %d", wins)
	}
	if got, err := s.ReadLock("shared/target"); err != nil || got.LockedBy == "zombie" {
		t.Errorf("stale lock was not replaced: %+v, %v", got, err)
	}
	if entries, _ := os.ReadDir(filepath.Join(s.baseDir, "locks")); len(entries) != 1 {
		t.Errorf("expected only the lock file to remain, found %d entries", len(entries))
	}
}
