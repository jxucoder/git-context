package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/git-context/internal/model"
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
  git ctx add --shared "Team standards"
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
	// Get title from args or flag
	title := addTitle
	if title == "" && len(args) > 0 {
		title = strings.Join(args, " ")
	}
	if title == "" {
		title = "Untitled"
	}
	
	// Get content from message flag, stdin, or editor
	content := addMessage
	
	if content == "" {
		// Check if stdin has data
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Reading from pipe
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
			content = string(data)
		} else {
			// Open editor
			var err error
			content, err = openEditor(title)
			if err != nil {
				return fmt.Errorf("editor failed: %w", err)
			}
		}
	}
	
	// Trim whitespace
	content = strings.TrimSpace(content)
	
	// Create memory entry
	author := model.GetAuthorShort()
	m := model.NewMemory(title, content, author, flagShared)
	m.Tags = addTags
	
	// Save
	storage := getStorage()
	if err := storage.WriteMemory(m); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}
	
	// Output
	storageType := "local"
	if flagShared {
		storageType = "shared"
	}
	fmt.Printf("Created (%s): %s\n", storageType, m.ID)
	
	return nil
}

func openEditor(title string) (string, error) {
	// Create temp file
	tmpfile, err := os.CreateTemp("", "git-ctx-*.md")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())
	
	// Write template
	template := fmt.Sprintf("# %s\n\n", title)
	tmpfile.WriteString(template)
	tmpfile.Close()
	
	// Open editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	
	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return "", err
	}
	
	// Read content
	file, err := os.Open(tmpfile.Name())
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	
	return strings.Join(lines, "\n"), nil
}

