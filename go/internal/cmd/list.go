package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/git-context/internal/model"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List context entries",
	Long: `List context entries from git.

By default, shows local entries. Use --shared for shared entries,
or --all for everything.

Examples:
  git ctx list                # Local entries
  git ctx list --shared       # Shared entries
  git ctx list --all          # Everything
  git ctx list --json         # JSON output`,
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	var memories []*model.Memory
	
	// Collect entries based on flags
	if flagAll || !flagShared {
		local, err := store.Local.ListMemories()
		if err != nil {
			return fmt.Errorf("failed to list local: %w", err)
		}
		for _, m := range local {
			m.Shared = false
			memories = append(memories, m)
		}
	}
	
	if flagAll || flagShared {
		shared, err := store.Shared.ListMemories()
		if err != nil {
			return fmt.Errorf("failed to list shared: %w", err)
		}
		for _, m := range shared {
			m.Shared = true
			memories = append(memories, m)
		}
	}
	
	// Output
	if flagJSON {
		return outputJSON(memories)
	}
	
	return outputTable(memories)
}

func outputJSON(memories []*model.Memory) error {
	data, err := json.MarshalIndent(memories, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputTable(memories []*model.Memory) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	
	fmt.Fprintln(w, "ID\tTITLE\tTYPE\tAUTHOR")
	fmt.Fprintln(w, "----\t-----\t----\t------")
	
	for _, m := range memories {
		typeStr := "[local]"
		if m.Shared {
			typeStr = "[shared]"
		}
		
		title := m.Title
		if len(title) > 45 {
			title = title[:42] + "..."
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.ID, title, typeStr, m.Author)
	}
	
	return w.Flush()
}


