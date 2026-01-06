package storage

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/git-context/internal/model"
)

// LocalStorage stores data in .git/context/ as plain files.
// This storage is private and never syncs.
type LocalStorage struct {
	baseDir string // .git/context/
}

// NewLocalStorage creates a new local storage instance.
func NewLocalStorage(gitDir string) (*LocalStorage, error) {
	baseDir := filepath.Join(gitDir, "context")
	
	// Create directories
	for _, subdir := range []string{"memory", "tasks", "locks"} {
		dir := filepath.Join(baseDir, subdir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create %s: %w", dir, err)
		}
	}
	
	return &LocalStorage{baseDir: baseDir}, nil
}

// Memory operations

func (s *LocalStorage) WriteMemory(m *model.Memory) error {
	dir := filepath.Join(s.baseDir, "memory", m.ID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// Write metadata
	meta := struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		Author    string    `json:"author"`
		Tags      []string  `json:"tags,omitempty"`
		CreatedAt string    `json:"createdAt"`
		UpdatedAt string    `json:"updatedAt"`
		Shared    bool      `json:"shared"`
	}{
		ID:        m.ID,
		Title:     m.Title,
		Author:    m.Author,
		Tags:      m.Tags,
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: m.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Shared:    m.Shared,
	}
	
	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), metaBytes, 0644); err != nil {
		return err
	}
	
	// Write content
	return os.WriteFile(filepath.Join(dir, "content.md"), []byte(m.Content), 0644)
}

func (s *LocalStorage) ReadMemory(id string) (*model.Memory, error) {
	dir := filepath.Join(s.baseDir, "memory", id)
	
	// Read metadata
	metaBytes, err := os.ReadFile(filepath.Join(dir, "meta.json"))
	if err != nil {
		return nil, err
	}
	
	var meta struct {
		ID        string   `json:"id"`
		Title     string   `json:"title"`
		Author    string   `json:"author"`
		Tags      []string `json:"tags,omitempty"`
		CreatedAt string   `json:"createdAt"`
		UpdatedAt string   `json:"updatedAt"`
		Shared    bool     `json:"shared"`
	}
	
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return nil, err
	}
	
	// Read content
	content, err := os.ReadFile(filepath.Join(dir, "content.md"))
	if err != nil {
		return nil, err
	}
	
	// Parse times
	createdAt, _ := parseTime(meta.CreatedAt)
	updatedAt, _ := parseTime(meta.UpdatedAt)
	
	return &model.Memory{
		ID:        meta.ID,
		Title:     meta.Title,
		Content:   string(content),
		Author:    meta.Author,
		Tags:      meta.Tags,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Shared:    meta.Shared,
	}, nil
}

func (s *LocalStorage) ListMemories() ([]*model.Memory, error) {
	dir := filepath.Join(s.baseDir, "memory")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	
	var memories []*model.Memory
	for _, entry := range entries {
		if entry.IsDir() {
			m, err := s.ReadMemory(entry.Name())
			if err == nil {
				memories = append(memories, m)
			}
		}
	}
	
	return memories, nil
}

func (s *LocalStorage) DeleteMemory(id string) error {
	dir := filepath.Join(s.baseDir, "memory", id)
	return os.RemoveAll(dir)
}

func (s *LocalStorage) SearchMemories(query string) ([]*model.Memory, error) {
	memories, err := s.ListMemories()
	if err != nil {
		return nil, err
	}
	
	var results []*model.Memory
	for _, m := range memories {
		if m.MatchesSearch(query) {
			results = append(results, m)
		}
	}
	
	return results, nil
}

// Task operations

func (s *LocalStorage) WriteTask(t *model.Task) error {
	dir := filepath.Join(s.baseDir, "tasks")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filepath.Join(dir, t.ID+".json"), data, 0644)
}

func (s *LocalStorage) ReadTask(id string) (*model.Task, error) {
	path := filepath.Join(s.baseDir, "tasks", id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var t model.Task
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	
	return &t, nil
}

func (s *LocalStorage) ListTasks() ([]*model.Task, error) {
	dir := filepath.Join(s.baseDir, "tasks")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	
	var tasks []*model.Task
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			id := strings.TrimSuffix(entry.Name(), ".json")
			t, err := s.ReadTask(id)
			if err == nil {
				tasks = append(tasks, t)
			}
		}
	}
	
	return tasks, nil
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
	path := filepath.Join(s.baseDir, "tasks", id+".json")
	return os.Remove(path)
}

// Lock operations

func (s *LocalStorage) WriteLock(l *model.Lock) error {
	dir := filepath.Join(s.baseDir, "locks")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	
	hash := hashTarget(l.Target)
	return os.WriteFile(filepath.Join(dir, hash+".json"), data, 0644)
}

func (s *LocalStorage) ReadLock(target string) (*model.Lock, error) {
	hash := hashTarget(target)
	path := filepath.Join(s.baseDir, "locks", hash+".json")
	
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var l model.Lock
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, err
	}
	
	return &l, nil
}

func (s *LocalStorage) ListLocks() ([]*model.Lock, error) {
	dir := filepath.Join(s.baseDir, "locks")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	
	var locks []*model.Lock
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			path := filepath.Join(dir, entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			
			var l model.Lock
			if err := json.Unmarshal(data, &l); err == nil {
				locks = append(locks, &l)
			}
		}
	}
	
	return locks, nil
}

func (s *LocalStorage) DeleteLock(target string) error {
	hash := hashTarget(target)
	path := filepath.Join(s.baseDir, "locks", hash+".json")
	return os.Remove(path)
}

// Helpers

func hashTarget(target string) string {
	h := sha256.Sum256([]byte(target))
	return fmt.Sprintf("%x", h[:8])
}

func parseTime(s string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05Z", s)
}

