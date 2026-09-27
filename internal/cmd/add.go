package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jxucoder/git-context/internal/model"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	addTitle   string
	addMessage string
	addTags    []string
)

var addCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Add a new context entry",
	Long: `Add a new context entry to git.

Opens an editor if no --message is provided. Content can also be piped via stdin.

Examples:
  git ctx add "Why JWT for auth"
  git ctx add --title "Decision" --message "We chose X because..."
  echo "content" | git ctx add --title "Note"
  git ctx add --title "Auth" --tag=security --tag=backend`,
	RunE: runAdd,
}

func init() {
	addCmd.Flags().StringVarP(&addTitle, "title", "t", "", "Entry title")
	addCmd.Flags().StringVarP(&addMessage, "message", "m", "", "Entry content (skips editor)")
	addCmd.Flags().StringArrayVar(&addTags, "tag", nil, "Tags for categorization")
}

func runAdd(cmd *cobra.Command, args []string) error {
	title := addTitle
	if title == "" && len(args) > 0 {
		title = strings.Join(args, " ")
	}
	if title == "" {
		title = "Untitled"
	}

	content := addMessage
	if content == "" {
		var err error
		if content, err = readContent(title, term.IsTerminal(int(os.Stdin.Fd()))); err != nil {
			return err
		}
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return fmt.Errorf("no content: pass -m, pipe text on stdin, or run in a terminal to use the editor")
	}

	m := model.NewMemory(title, content, model.GetAuthorShort(), flagShared)
	m.Tags = addTags

	b := targetBackend()
	if err := b.WriteMemory(m); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Printf("Created (%s): %s\n", b.name, m.ID)
	return nil
}

// readContent takes the entry body from stdin when it is piped or redirected,
// and from the editor when running interactively.
func readContent(title string, interactive bool) (string, error) {
	if !interactive {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read stdin: %w", err)
		}
		return string(data), nil
	}

	template := fmt.Sprintf("# %s\n\n", title)
	content, err := openEditor(template)
	if err != nil {
		return "", err
	}
	if content == strings.TrimSpace(template) {
		return "", fmt.Errorf("aborted: nothing was added below the title")
	}
	return content, nil
}
