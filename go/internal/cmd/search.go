package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/git-context/internal/model"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search context entries",
	Long: `Search context entries by title and content.

Searches both local and shared storage by default.

Examples:
  git ctx search "auth"
  git ctx search "JWT tokens"`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]
	
	var results []*model.Memory
	
	// Search local
	local, err := store.Local.SearchMemories(query)
	if err == nil {
		for _, m := range local {
			m.Shared = false
			results = append(results, m)
		}
	}
	
	// Search shared
	shared, err := store.Shared.SearchMemories(query)
	if err == nil {
		for _, m := range shared {
			m.Shared = true
			results = append(results, m)
		}
	}
	
	if len(results) == 0 {
		fmt.Printf("No results for: %s\n", query)
		return nil
	}
	
	// Output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	
	fmt.Fprintf(w, "Found %d results for \"%s\":\n\n", len(results), query)
	fmt.Fprintln(w, "ID\tTITLE\tTYPE")
	fmt.Fprintln(w, "----\t-----\t----")
	
	for _, m := range results {
		typeStr := "[local]"
		if m.Shared {
			typeStr = "[shared]"
		}
		
		title := m.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\n", m.ID, title, typeStr)
	}
	
	return w.Flush()
}

