// Package model defines the data structures for git-context.
package model

import (
	"strings"
	"time"
)

// Memory represents a context entry (notes, decisions, plans).
type Memory struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content,omitempty"`
	Author    string    `json:"author"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Shared    bool      `json:"shared"`
}

// NewMemory creates a new memory entry with generated ID.
func NewMemory(title, content, author string, shared bool) *Memory {
	now := time.Now().UTC()
	return &Memory{
		ID:        GenerateID(),
		Title:     title,
		Content:   content,
		Author:    author,
		CreatedAt: now,
		UpdatedAt: now,
		Shared:    shared,
	}
}

// MatchesSearch returns true if the memory matches the search query.
func (m *Memory) MatchesSearch(query string) bool {
	query = strings.ToLower(query)
	return strings.Contains(strings.ToLower(m.Title), query) ||
		strings.Contains(strings.ToLower(m.Content), query)
}

