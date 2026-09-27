// git-ctx: Store coding context in git
//
// A CLI tool for managing notes, decisions, tasks and locks within git
// repositories. Entries live in .git/context and are private to the clone;
// shared storage that syncs with push/pull is not available yet.
//
// Usage:
//
//	git ctx add "Why I chose JWT"      # Add context
//	git ctx list                        # List entries
//	git ctx task add "Implement auth"   # Create task
//	git ctx lock src/auth/              # Lock a path
package main

import (
	"fmt"
	"os"

	"github.com/jxucoder/git-context/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
