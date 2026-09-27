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

	for _, b := range allBackends() {
		err := b.DeleteMemory(id)
		if err == nil {
			fmt.Printf("Removed (%s): %s\n", b.name, id)
			return nil
		}
		if !unavailable(err) {
			return fmt.Errorf("failed to remove: %w", err)
		}
	}

	return fmt.Errorf("not found: %s", id)
}
