package model

import (
	"errors"
	"fmt"
	"time"
)

// TaskStatus represents the current state of a task.
type TaskStatus string

const (
	TaskOpen    TaskStatus = "open"
	TaskClaimed TaskStatus = "claimed"
	TaskDone    TaskStatus = "done"
)

var (
	// ErrTaskDone is returned when claiming or dropping a completed task.
	ErrTaskDone = errors.New("task is already done")
	// ErrNotOwner is returned when dropping a task owned by someone else.
	ErrNotOwner = errors.New("task is not owned by you")
)

// ClaimedError is returned when claiming a task that someone else owns.
type ClaimedError struct {
	Owner string
}

func (e *ClaimedError) Error() string {
	return "already claimed by " + e.Owner
}

// Comment represents a comment on a task.
type Comment struct {
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// Task represents a work item for tracking and coordination.
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      TaskStatus `json:"status"`
	Owner       string     `json:"owner,omitempty"`
	CreatedBy   string     `json:"createdBy"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DoneAt      *time.Time `json:"doneAt,omitempty"`
	BlockedBy   []string   `json:"blockedBy,omitempty"`
	Blocks      []string   `json:"blocks,omitempty"`
	Comments    []Comment  `json:"comments,omitempty"`
	Shared      bool       `json:"shared"`
}

// NewTask creates a new task with generated ID.
func NewTask(title, description, author string, shared bool) *Task {
	now := time.Now().UTC()
	return &Task{
		ID:          "task-" + GenerateID(),
		Title:       title,
		Description: description,
		Status:      TaskOpen,
		CreatedBy:   author,
		CreatedAt:   now,
		UpdatedAt:   now,
		Shared:      shared,
	}
}

// Claim assigns the task to owner. Re-claiming a task you already own is a
// no-op, so a retried claim succeeds; done tasks cannot be claimed.
func (t *Task) Claim(owner string) error {
	switch {
	case t.Status == TaskDone:
		return ErrTaskDone
	case t.Status == TaskClaimed && t.Owner != owner:
		return &ClaimedError{Owner: t.Owner}
	case t.Status == TaskClaimed:
		return nil
	}
	t.Owner = owner
	t.Status = TaskClaimed
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// Drop releases ownership. Only the owner can drop, and done tasks stay done.
func (t *Task) Drop(owner string) error {
	if t.Status == TaskDone {
		return ErrTaskDone
	}
	if t.Owner != owner {
		return fmt.Errorf("%w (owner: %s)", ErrNotOwner, t.OwnerLabel())
	}
	t.Owner = ""
	t.Status = TaskOpen
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// Done marks the task as complete. Marking a done task again is a no-op.
func (t *Task) Done() {
	if t.Status == TaskDone {
		return
	}
	now := time.Now().UTC()
	t.Status = TaskDone
	t.DoneAt = &now
	t.UpdatedAt = now
}

// AddComment adds a comment to the task.
func (t *Task) AddComment(author, content string) {
	now := time.Now().UTC()
	t.Comments = append(t.Comments, Comment{
		Author:    author,
		Content:   content,
		CreatedAt: now,
	})
	t.UpdatedAt = now
}

// OwnerLabel returns the owner, or "-" for unowned tasks.
func (t *Task) OwnerLabel() string {
	if t.Owner == "" {
		return "-"
	}
	return t.Owner
}
