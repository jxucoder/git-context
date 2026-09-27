package model

import (
	"errors"
	"testing"
	"time"
)

func TestGenerateID_Format(t *testing.T) {
	id := GenerateID()
	if len(id) != 8 {
		t.Errorf("expected 8-char hex ID, got %q (len=%d)", id, len(id))
	}
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("ID %q contains non-hex character %q", id, c)
		}
	}
}

func TestGenerateID_Unique(t *testing.T) {
	// 100 draws from a 32-bit space collide with probability ~1e-6, which is
	// low enough not to make this flaky while still catching a broken RNG.
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GenerateID()
		if seen[id] {
			t.Fatalf("duplicate ID generated: %s", id)
		}
		seen[id] = true
	}
}

func TestMemory_MatchesSearch(t *testing.T) {
	m := NewMemory("JWT Authentication", "We chose JWT because it is stateless", "alice", false)

	tests := []struct {
		query string
		want  bool
	}{
		{"jwt", true},
		{"JWT", true},
		{"stateless", true},
		{"oauth", false},
		{"", true},
	}

	for _, tt := range tests {
		if got := m.MatchesSearch(tt.query); got != tt.want {
			t.Errorf("MatchesSearch(%q) = %v, want %v", tt.query, got, tt.want)
		}
	}
}

func TestTask_StateMachine(t *testing.T) {
	task := NewTask("Fix bug", "critical issue", "alice", false)
	if task.Status != TaskOpen {
		t.Fatalf("expected TaskOpen, got %s", task.Status)
	}

	if err := task.Claim("alice"); err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if task.Status != TaskClaimed || task.Owner != "alice" {
		t.Errorf("after Claim: status=%s owner=%s", task.Status, task.Owner)
	}

	// Re-claiming your own task is a no-op so retries succeed.
	if err := task.Claim("alice"); err != nil {
		t.Errorf("re-claim by owner: %v", err)
	}

	var claimed *ClaimedError
	if err := task.Claim("bob"); !errors.As(err, &claimed) || claimed.Owner != "alice" {
		t.Errorf("claim by bob: got %v, want ClaimedError{alice}", err)
	}
	if err := task.Drop("bob"); !errors.Is(err, ErrNotOwner) {
		t.Errorf("drop by bob: got %v, want ErrNotOwner", err)
	}

	if err := task.Drop("alice"); err != nil {
		t.Fatalf("Drop: %v", err)
	}
	if task.Status != TaskOpen || task.Owner != "" {
		t.Errorf("after Drop: status=%s owner=%q", task.Status, task.Owner)
	}

	task.Done()
	if task.Status != TaskDone || task.DoneAt == nil {
		t.Fatalf("after Done: status=%s doneAt=%v", task.Status, task.DoneAt)
	}
	doneAt := *task.DoneAt

	if err := task.Claim("alice"); !errors.Is(err, ErrTaskDone) {
		t.Errorf("claim after done: got %v, want ErrTaskDone", err)
	}
	if err := task.Drop("alice"); !errors.Is(err, ErrTaskDone) {
		t.Errorf("drop after done: got %v, want ErrTaskDone", err)
	}
	task.Done()
	if !task.DoneAt.Equal(doneAt) {
		t.Errorf("Done is not idempotent: doneAt changed from %v to %v", doneAt, *task.DoneAt)
	}
}

func TestTask_AddComment(t *testing.T) {
	task := NewTask("Implement feature", "", "bob", false)
	task.AddComment("alice", "looks good")
	task.AddComment("bob", "will do")

	if len(task.Comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(task.Comments))
	}
	if task.Comments[0].Author != "alice" || task.Comments[0].Content != "looks good" {
		t.Errorf("unexpected first comment: %+v", task.Comments[0])
	}
}

func TestLock_ExpiryAndOwnership(t *testing.T) {
	lock := NewLock("src/auth/", "carol")

	if lock.IsExpired() {
		t.Error("new lock should not be expired")
	}
	if !lock.IsOwnedBy("carol") || lock.IsOwnedBy("dave") {
		t.Error("lock ownership is wrong")
	}
	if lock.ExpiresAt.Sub(lock.LockedAt) != DefaultLockExpiry {
		t.Errorf("expiry is %v after lock time, want %v", lock.ExpiresAt.Sub(lock.LockedAt), DefaultLockExpiry)
	}

	lock.ExpiresAt = time.Now().UTC().Add(-time.Second)
	if !lock.IsExpired() {
		t.Error("lock past its expiry should be expired")
	}
}

func TestNormalizeTarget(t *testing.T) {
	tests := map[string]string{
		"src/auth":     "src/auth",
		"src/auth/":    "src/auth",
		"./src/auth":   "src/auth",
		" src/auth ":   "src/auth",
		"src//auth/..": "src",
		"task-abc":     "task-abc",
		"":             "",
		"   ":          "",
	}
	for in, want := range tests {
		if got := NormalizeTarget(in); got != want {
			t.Errorf("NormalizeTarget(%q) = %q, want %q", in, got, want)
		}
	}
}
