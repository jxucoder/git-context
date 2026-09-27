package cmd

import (
	"fmt"
	"time"

	"github.com/jxucoder/git-context/internal/model"
	"github.com/spf13/cobra"
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
	m, b, err := findMemory(args[0])
	if err != nil {
		return err
	}

	if flagJSON {
		return printJSON(m)
	}

	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Printf("  %s\n", m.Title)
	fmt.Printf("  by %s • %s • [%s]\n", m.Author, m.CreatedAt.UTC().Format(time.RFC3339), b.name)
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println(m.Content)

	return nil
}

// findMemory looks the id up in every backend, local first.
func findMemory(id string) (*model.Memory, backend, error) {
	for _, b := range allBackends() {
		m, err := b.ReadMemory(id)
		if err == nil {
			m.Shared = b.shared
			return m, b, nil
		}
		if !unavailable(err) {
			return nil, b, err
		}
	}
	return nil, backend{}, fmt.Errorf("not found: %s", id)
}
