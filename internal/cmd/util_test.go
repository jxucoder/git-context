package cmd

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		in   string
		max  int
		want string
	}{
		{"short", 10, "short"},
		{"exactly ten", 11, "exactly ten"},
		{"a much longer title here", 10, "a much ..."},
		{"日本語のタイトルがとても長い場合", 8, "日本語のタ..."},
	}
	for _, tt := range tests {
		if got := truncate(tt.in, tt.max); got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.in, tt.max, got, tt.want)
		}
	}
}

func TestNeedsStorage(t *testing.T) {
	root := &cobra.Command{Use: "git-ctx"}
	completion := &cobra.Command{Use: "completion"}
	bash := &cobra.Command{Use: "bash"}
	completion.AddCommand(bash)
	list := &cobra.Command{Use: "list"}
	root.AddCommand(completion, list, &cobra.Command{Use: "help"}, &cobra.Command{Use: cobra.ShellCompRequestCmd})

	for _, tt := range []struct {
		cmd  *cobra.Command
		want bool
	}{
		{root, true},
		{list, true},
		{completion, false},
		{bash, false},
	} {
		if got := needsStorage(tt.cmd); got != tt.want {
			t.Errorf("needsStorage(%s) = %v, want %v", tt.cmd.Name(), got, tt.want)
		}
	}
	for _, name := range []string{"help", cobra.ShellCompRequestCmd} {
		c, _, err := root.Find([]string{name})
		if err != nil {
			t.Fatal(err)
		}
		if needsStorage(c) {
			t.Errorf("needsStorage(%s) = true, want false", name)
		}
	}
}

func TestEditorCommand_SplitsArguments(t *testing.T) {
	t.Setenv("GIT_EDITOR", "code --wait")
	if got := editorCommand(); !reflect.DeepEqual(got, []string{"code", "--wait"}) {
		t.Errorf("editorCommand() = %v, want [code --wait]", got)
	}

	t.Setenv("GIT_EDITOR", "")
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "  emacs -nw  ")
	got := editorCommand()
	// core.editor may be set on the developer's machine; only assert when the
	// fallback chain reached $EDITOR.
	if got[0] == "emacs" && !reflect.DeepEqual(got, []string{"emacs", "-nw"}) {
		t.Errorf("editorCommand() = %v, want [emacs -nw]", got)
	}
}
