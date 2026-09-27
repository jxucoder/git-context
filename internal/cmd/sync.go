package cmd

import (
	"fmt"

	"github.com/jxucoder/git-context/internal/storage"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push shared context to remote (not available yet)",
	Long: `Push shared context entries to the remote repository.

Shared storage is not implemented yet, so this command currently fails.`,
	RunE: runPush,
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull shared context from remote (not available yet)",
	Long: `Pull shared context entries from the remote repository.

Shared storage is not implemented yet, so this command currently fails.`,
	RunE: runPull,
}

func runPush(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("%w: push is not available in this version", storage.ErrNotImplemented)
}

func runPull(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("%w: pull is not available in this version", storage.ErrNotImplemented)
}
