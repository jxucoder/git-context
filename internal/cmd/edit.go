package cmd

import (
	"fmt"
	"strings"
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

	m, b, err := findMemory(id)
	if err != nil {
		return err
	}

	content, err := openEditor(m.Content)
	if err != nil {
		return err
	}
	if content == strings.TrimSpace(m.Content) {
		fmt.Println("No changes")
		return nil
	}

	m.Content = content
	m.UpdatedAt = time.Now().UTC()

	if err := b.WriteMemory(m); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Printf("Updated (%s): %s\n", b.name, id)
	return nil
}
