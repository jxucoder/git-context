package model

import (
	"strings"
	"time"
)

// TaskStatus represents the current state of a task.
type TaskStatus string

const (
	TaskOpen    TaskStatus = "open"
	TaskClaimed TaskStatus = "claimed"
	TaskDone    TaskStatus = "done"
)

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

// Claim assigns the task to the given owner.
func (t *Task) Claim(owner string) {
	t.Owner = owner
	t.Status = TaskClaimed
	t.UpdatedAt = time.Now().UTC()
}

// Drop releases the task ownership.
func (t *Task) Drop() {
	t.Owner = ""
	t.Status = TaskOpen
	t.UpdatedAt = time.Now().UTC()
}

// Done marks the task as complete.
func (t *Task) Done() {
	now := time.Now().UTC()
	t.Status = TaskDone
	t.DoneAt = &now
	t.UpdatedAt = now
}

// AddComment adds a comment to the task.
func (t *Task) AddComment(author, content string) {
	t.Comments = append(t.Comments, Comment{
		Author:    author,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	})
	t.UpdatedAt = time.Now().UTC()
}

// IsBlocked returns true if any blocking tasks are not done.
func (t *Task) IsBlocked(tasks map[string]*Task) bool {
	for _, id := range t.BlockedBy {
		if blocker, ok := tasks[id]; ok {
			if blocker.Status != TaskDone {
				return true
			}
		}
	}
	return false
}

// MatchesSearch returns true if the task matches the search query.
func (t *Task) MatchesSearch(query string) bool {
	query = strings.ToLower(query)
	return strings.Contains(strings.ToLower(t.Title), query) ||
		strings.Contains(strings.ToLower(t.Description), query)
}


