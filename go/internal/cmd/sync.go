package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push shared context to remote",
	Long: `Push shared context entries to the remote repository.

Only shared entries are pushed (created with --shared flag).

Examples:
  git ctx push`,
	RunE: runPush,
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull shared context from remote",
	Long: `Pull shared context entries from the remote repository.

Examples:
  git ctx pull`,
	RunE: runPull,
}

func runPush(cmd *cobra.Command, args []string) error {
	// TODO: Implement git refs push
	// For now, using the Bash version's approach would work:
	// git push origin 'refs/context/*:refs/context/*'
	
	fmt.Println("Push not yet implemented in Go version.")
	fmt.Println("Use the Bash version for push/pull, or run:")
	fmt.Println("  git push origin 'refs/context/*:refs/context/*'")
	
	return nil
}

func runPull(cmd *cobra.Command, args []string) error {
	// TODO: Implement git refs pull
	// For now, using the Bash version's approach would work:
	// git fetch origin 'refs/context/*:refs/context/*'
	
	fmt.Println("Pull not yet implemented in Go version.")
	fmt.Println("Use the Bash version for push/pull, or run:")
	fmt.Println("  git fetch origin 'refs/context/*:refs/context/*'")
	
	return nil
}


