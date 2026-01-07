// git-ctx: Store coding context in git
//
// A CLI tool for managing notes, decisions, and tasks within git repositories.
// Local by default, use --shared to sync with team via push/pull.
//
// Usage:
//
//	git ctx add "Why I chose JWT"      # Add context
//	git ctx list                        # List entries
//	git ctx task add "Implement auth"   # Create task
//	git ctx push                        # Push shared to remote
package main

import (
	"fmt"
	"os"

	"github.com/user/git-context/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}


