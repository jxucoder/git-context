package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <id>",
	Short: "Remove a context entry",
	Long: `Remove a context entry from git.

Searches both local and shared storage.

Examples:
  git ctx rm abc12345`,
	Args: cobra.ExactArgs(1),
	RunE: runRm,
}

func runRm(cmd *cobra.Command, args []string) error {
	id := args[0]
	
	// Try local first
	err := store.Local.DeleteMemory(id)
	if err == nil {
		fmt.Printf("Removed (local): %s\n", id)
		return nil
	}
	
	// Try shared
	err = store.Shared.DeleteMemory(id)
	if err == nil {
		fmt.Printf("Removed (shared): %s\n", id)
		return nil
	}
	
	return fmt.Errorf("not found: %s", id)
}


