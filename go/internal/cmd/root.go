// Package cmd implements the CLI commands for git-context.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/git-context/internal/storage"
)

var (
	// Global flags
	flagShared bool
	flagAll    bool
	flagJSON   bool
	
	// Storage instances
	store *storage.MultiStorage
)

var rootCmd = &cobra.Command{
	Use:   "git-ctx",
	Short: "Store coding context in git",
	Long: `git-context: Simple context storage for vibe coding.

Store notes, decisions, and tasks in git. Local by default, 
use --shared to sync with team via push/pull.

Examples:
  git ctx add "Why I chose JWT"          # Add context
  git ctx list                            # List local entries
  git ctx task add "Implement auth"       # Create task
  git ctx push                            # Push shared to remote`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip init for help
		if cmd.Name() == "help" {
			return nil
		}
		
		// Find git directory
		gitDir, err := findGitDir()
		if err != nil {
			return fmt.Errorf("not a git repository")
		}
		
		// Initialize storage
		store, err = storage.NewMultiStorage(gitDir)
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}
		
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&flagShared, "shared", "s", false, "Use shared storage (syncs with push/pull)")
	rootCmd.PersistentFlags().BoolVarP(&flagAll, "all", "a", false, "Show both local and shared")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output as JSON")
	
	// Add subcommands
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(lockCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(pullCmd)
}

// findGitDir finds the .git directory.
func findGitDir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getStorage returns the appropriate storage based on flags.
func getStorage() storage.Storage {
	if flagShared {
		return store.Shared
	}
	return store.Local
}

// die prints an error and exits.
func die(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}


