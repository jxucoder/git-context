package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a context entry",
	Long: `Edit a context entry in your default editor.

Searches both local and shared storage.

Examples:
  git ctx edit abc12345`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

func runEdit(cmd *cobra.Command, args []string) error {
	id := args[0]
	
	// Find entry
	m, storageType := findMemory(id)
	if m == nil {
		return fmt.Errorf("not found: %s", id)
	}
	
	// Create temp file with content
	tmpfile, err := os.CreateTemp("", "git-ctx-*.md")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())
	
	tmpfile.WriteString(m.Content)
	tmpfile.Close()
	
	// Open editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	
	editCmd := exec.Command(editor, tmpfile.Name())
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr
	
	if err := editCmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}
	
	// Read new content
	newContent, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return err
	}
	
	// Update entry
	m.Content = string(newContent)
	m.UpdatedAt = time.Now().UTC()
	
	// Save to correct storage
	var saveErr error
	if storageType == "local" {
		saveErr = store.Local.WriteMemory(m)
	} else {
		saveErr = store.Shared.WriteMemory(m)
	}
	
	if saveErr != nil {
		return fmt.Errorf("failed to save: %w", saveErr)
	}
	
	fmt.Printf("Updated: %s\n", id)
	return nil
}


