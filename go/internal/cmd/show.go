package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/git-context/internal/model"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a context entry",
	Long: `Show the full content of a context entry.

Searches both local and shared storage.

Examples:
  git ctx show abc12345
  git ctx show abc12345 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runShow,
}

func runShow(cmd *cobra.Command, args []string) error {
	id := args[0]
	
	// Try local first, then shared
	m, storageType := findMemory(id)
	if m == nil {
		return fmt.Errorf("not found: %s", id)
	}
	
	if flagJSON {
		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
	
	// Pretty print
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Printf("  %s\n", m.Title)
	fmt.Printf("  by %s • %s • [%s]\n", m.Author, m.CreatedAt.Format("2006-01-02T15:04:05Z"), storageType)
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println(m.Content)
	
	return nil
}

func findMemory(id string) (*model.Memory, string) {
	// Try local
	m, err := store.Local.ReadMemory(id)
	if err == nil && m != nil {
		return m, "local"
	}
	
	// Try shared
	m, err = store.Shared.ReadMemory(id)
	if err == nil && m != nil {
		return m, "shared"
	}
	
	return nil, ""
}


