package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/jxucoder/git-context/internal/storage"
	"github.com/spf13/cobra"
)

// version is set at build time with
// -ldflags "-X github.com/jxucoder/git-context/internal/cmd.version=1.2.3".
// Builds made with `go install module@vX.Y.Z` fall back to the module version
// the Go toolchain records.
var version = "dev"

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

Store notes, decisions, tasks and locks in git. Entries live in .git/context
and are private to this clone. Shared storage that syncs with the team
(--shared, push, pull) is not available yet.

Examples:
  git ctx add "Why we use JWT" -m "..."   # Add context
  git ctx task add "Implement auth"       # Create task
  git ctx lock src/auth/                  # Lock a path while working on it`,
	// main prints the error once. Usage is only shown for argument errors,
	// which cobra reports before PersistentPreRunE runs.
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.Root().SilenceUsage = true

		if !needsStorage(cmd) {
			return nil
		}
		if flagShared {
			return fmt.Errorf("%w: the --shared flag is not available in this version", storage.ErrNotImplemented)
		}
		if flagAll {
			warnf("shared storage is not implemented yet; --all shows local entries only")
		}

		gitDir, err := findGitDir()
		if err != nil {
			return fmt.Errorf("not a git repository")
		}

		store, err = storage.NewMultiStorage(gitDir)
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		return nil
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = strings.TrimPrefix(info.Main.Version, "v")
		}
	}
	rootCmd.Version = version

	rootCmd.PersistentFlags().BoolVarP(&flagShared, "shared", "s", false, "Use shared storage (not available yet)")
	rootCmd.PersistentFlags().BoolVarP(&flagAll, "all", "a", false, "Show both local and shared")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output as JSON")

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

// needsStorage reports whether cmd reads or writes storage. Help, completion
// and version output must work outside a git repository.
func needsStorage(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		switch c.Name() {
		case "help", "completion", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
			return false
		}
	}
	return true
}

// findGitDir returns the absolute path of the repository's common git
// directory, so every worktree of a clone shares one context store.
func findGitDir() (string, error) {
	output, err := exec.Command("git", "rev-parse", "--git-common-dir").Output()
	if err != nil {
		return "", err
	}
	return filepath.Abs(strings.TrimSpace(string(output)))
}
