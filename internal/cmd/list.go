package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/jxucoder/git-context/internal/model"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List context entries",
	Long: `List context entries from git, newest first.

Shows local entries. --all will include shared entries once shared storage
is implemented.

Examples:
  git ctx list                # Local entries
  git ctx list --json         # JSON output`,
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	memories, err := collectMemories(selectedBackends(), backend.ListMemories)
	if err != nil {
		return err
	}

	if flagJSON {
		return printJSON(memories)
	}

	return printMemoryTable(memories)
}

// collectMemories merges the entries list returns for every backend, newest first.
func collectMemories(backends []backend, list func(backend) ([]*model.Memory, error)) ([]*model.Memory, error) {
	memories, err := collect(backends, list, func(m *model.Memory, b backend) { m.Shared = b.shared })
	if err != nil {
		return nil, err
	}

	slices.SortFunc(memories, func(a, b *model.Memory) int {
		if c := b.CreatedAt.Compare(a.CreatedAt); c != 0 {
			return c
		}
		return strings.Compare(a.ID, b.ID)
	})

	return memories, nil
}

func printMemoryTable(memories []*model.Memory) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "ID\tTITLE\tTYPE\tAUTHOR")
	fmt.Fprintln(w, "----\t-----\t----\t------")

	for _, m := range memories {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.ID, truncate(m.Title, 45), typeLabel(m.Shared), m.Author)
	}

	return w.Flush()
}
