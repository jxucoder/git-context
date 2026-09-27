package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func writeScript(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestReadContent_Editor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses sh scripts as editors")
	}

	// The editor receives its configured arguments followed by the file.
	appender := writeScript(t, "ed.sh", `for f; do :; done
[ "$1" = "--flag" ] || exit 3
echo "typed body" >> "$f"
`)
	t.Setenv("GIT_EDITOR", appender+" --flag")

	got, err := readContent("My title", true)
	if err != nil {
		t.Fatalf("readContent: %v", err)
	}
	if want := "# My title\n\ntyped body"; got != want {
		t.Errorf("readContent = %q, want %q", got, want)
	}

	// Leaving the template untouched aborts instead of saving "# My title".
	t.Setenv("GIT_EDITOR", writeScript(t, "noop.sh", "exit 0\n"))
	if _, err := readContent("My title", true); err == nil || !strings.Contains(err.Error(), "aborted") {
		t.Errorf("unchanged template: got %v, want an aborted error", err)
	}

	// A failing editor is reported, not treated as empty content.
	t.Setenv("GIT_EDITOR", writeScript(t, "fail.sh", "exit 7\n"))
	if _, err := readContent("My title", true); err == nil || !strings.Contains(err.Error(), "failed") {
		t.Errorf("failing editor: got %v, want a failed error", err)
	}
}
