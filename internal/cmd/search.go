package cmd

import (
	"fmt"

	"github.com/jxucoder/git-context/internal/model"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search context entries",
	Long: `Search context entries by title and content.

Searches both local and shared storage.

Examples:
  git ctx search "auth"
  git ctx search "JWT tokens" --json`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	results, err := collectMemories(allBackends(), func(b backend) ([]*model.Memory, error) {
		return b.SearchMemories(query)
	})
	if err != nil {
		return err
	}

	if flagJSON {
		return printJSON(results)
	}

	if len(results) == 0 {
		fmt.Printf("No results for: %s\n", query)
		return nil
	}

	fmt.Printf("Found %d results for %q:\n\n", len(results), query)
	return printMemoryTable(results)
}
