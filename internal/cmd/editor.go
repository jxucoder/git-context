package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jxucoder/git-context/internal/model"
)

// openEditor writes initial to a temporary file, opens it in the user's
// editor and returns the edited text with surrounding whitespace trimmed.
func openEditor(initial string) (string, error) {
	tmpfile, err := os.CreateTemp("", "git-ctx-*.md")
	if err != nil {
		return "", err
	}
	path := tmpfile.Name()
	defer os.Remove(path)

	if _, err := tmpfile.WriteString(initial); err != nil {
		tmpfile.Close()
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}

	editor := editorCommand()
	editorCmd := exec.Command(editor[0], append(editor[1:], path)...)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	if err := editorCmd.Run(); err != nil {
		return "", fmt.Errorf("editor %q failed: %w", editor[0], err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// editorCommand resolves the editor the way git does: GIT_EDITOR, then
// core.editor, VISUAL and EDITOR, falling back to vim. The value may carry
// arguments, e.g. "code --wait".
func editorCommand() []string {
	candidates := []string{
		os.Getenv("GIT_EDITOR"),
		model.GitConfig("core.editor"),
		os.Getenv("VISUAL"),
		os.Getenv("EDITOR"),
	}
	for _, c := range candidates {
		if fields := strings.Fields(c); len(fields) > 0 {
			return fields
		}
	}
	return []string{"vim"}
}
