package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/jxucoder/git-context/internal/model"
	"github.com/spf13/cobra"
)

var (
	taskDescription string
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
	Long:  `Create and manage tasks for tracking work.`,
}

var taskAddCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Create a new task",
	Long: `Create a new task for tracking work.

Examples:
  git ctx task add "Implement auth"
  git ctx task add "Setup database" -d "PostgreSQL schema with users table"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runTaskAdd,
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long: `List tasks, newest first.

Examples:
  git ctx task list           # Local tasks
  git ctx task list --json    # JSON output`,
	RunE: runTaskList,
}

var taskShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show task details",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskShow,
}

var taskClaimCmd = &cobra.Command{
	Use:   "claim <id>",
	Short: "Claim a task (take ownership)",
	Long: `Claim a task. Claiming a task you already own succeeds, so a retried
claim is safe; a task claimed by someone else or already done is refused.`,
	Args: cobra.ExactArgs(1),
	RunE: runTaskClaim,
}

var taskDropCmd = &cobra.Command{
	Use:   "drop <id>",
	Short: "Drop a task (release ownership)",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskDrop,
}

var taskDoneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark task as complete",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskDone,
}

var taskCommentCmd = &cobra.Command{
	Use:   "comment <id> <message>",
	Short: "Add a comment to a task",
	Args:  cobra.ExactArgs(2),
	RunE:  runTaskComment,
}

func init() {
	taskCmd.AddCommand(taskAddCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskShowCmd)
	taskCmd.AddCommand(taskClaimCmd)
	taskCmd.AddCommand(taskDropCmd)
	taskCmd.AddCommand(taskDoneCmd)
	taskCmd.AddCommand(taskCommentCmd)

	taskAddCmd.Flags().StringVarP(&taskDescription, "description", "d", "", "Task description")
}

func runTaskAdd(cmd *cobra.Command, args []string) error {
	title := strings.Join(args, " ")
	t := model.NewTask(title, taskDescription, model.GetAuthorShort(), flagShared)

	b := targetBackend()
	if err := b.WriteTask(t); err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Printf("Created (%s): %s\n", b.name, t.ID)
	return nil
}

func runTaskList(cmd *cobra.Command, args []string) error {
	tasks, err := collectTasks(selectedBackends())
	if err != nil {
		return err
	}

	if flagJSON {
		return printJSON(tasks)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tTYPE\tOWNER")
	fmt.Fprintln(w, "----\t-----\t------\t----\t-----")

	for _, t := range tasks {
		fmt.Fprintf(w, "%s\t%s\t[%s]\t%s\t%s\n", t.ID, truncate(t.Title, 35), t.Status, typeLabel(t.Shared), t.OwnerLabel())
	}

	return w.Flush()
}

// collectTasks merges the tasks of every backend, newest first.
func collectTasks(backends []backend) ([]*model.Task, error) {
	tasks, err := collect(backends, backend.ListTasks, func(t *model.Task, b backend) { t.Shared = b.shared })
	if err != nil {
		return nil, err
	}

	slices.SortFunc(tasks, func(a, b *model.Task) int {
		if c := b.CreatedAt.Compare(a.CreatedAt); c != 0 {
			return c
		}
		return strings.Compare(a.ID, b.ID)
	})

	return tasks, nil
}

func runTaskShow(cmd *cobra.Command, args []string) error {
	t, b, err := findTask(args[0])
	if err != nil {
		return err
	}

	if flagJSON {
		return printJSON(t)
	}

	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Printf("  %s\n", t.Title)
	fmt.Printf("  Status: %s • Type: %s • Created by: %s\n", t.Status, b.name, t.CreatedBy)
	fmt.Println("════════════════════════════════════════════════════════════")

	if t.Description != "" {
		fmt.Println()
		fmt.Println(t.Description)
	}

	if t.Owner != "" {
		fmt.Printf("\nOwner: %s\n", t.Owner)
	}

	if len(t.BlockedBy) > 0 {
		fmt.Printf("\nBlocked by: %s\n", strings.Join(t.BlockedBy, ", "))
	}

	if len(t.Comments) > 0 {
		fmt.Println("\nComments:")
		for _, c := range t.Comments {
			fmt.Printf("  [%s] %s: %s\n", c.CreatedAt.Format("2006-01-02"), c.Author, c.Content)
		}
	}

	return nil
}

func runTaskClaim(cmd *cobra.Command, args []string) error {
	id := args[0]
	author := model.GetAuthorShort()

	if _, err := updateTask(id, func(t *model.Task) error { return t.Claim(author) }); err != nil {
		return err
	}

	fmt.Printf("Claimed: %s\n", id)
	return nil
}

func runTaskDrop(cmd *cobra.Command, args []string) error {
	id := args[0]
	author := model.GetAuthorShort()

	if _, err := updateTask(id, func(t *model.Task) error { return t.Drop(author) }); err != nil {
		return err
	}

	fmt.Printf("Dropped: %s\n", id)
	return nil
}

func runTaskDone(cmd *cobra.Command, args []string) error {
	id := args[0]

	if _, err := updateTask(id, func(t *model.Task) error { t.Done(); return nil }); err != nil {
		return err
	}

	fmt.Printf("Done: %s\n", id)
	return nil
}

func runTaskComment(cmd *cobra.Command, args []string) error {
	id := args[0]
	message := args[1]
	author := model.GetAuthorShort()

	if _, err := updateTask(id, func(t *model.Task) error { t.AddComment(author, message); return nil }); err != nil {
		return err
	}

	fmt.Printf("Comment added to: %s\n", id)
	return nil
}

// findTask looks the id up in every backend, local first.
func findTask(id string) (*model.Task, backend, error) {
	for _, b := range allBackends() {
		t, err := b.ReadTask(id)
		if err == nil {
			t.Shared = b.shared
			return t, b, nil
		}
		if !unavailable(err) {
			return nil, b, err
		}
	}
	return nil, backend{}, fmt.Errorf("not found: %s", id)
}

// updateTask applies fn to the task in whichever backend stores it and saves
// the result. Errors returned by fn are passed through unchanged.
func updateTask(id string, fn func(*model.Task) error) (backend, error) {
	for _, b := range allBackends() {
		err := b.UpdateTask(id, fn)
		if err == nil {
			return b, nil
		}
		if !unavailable(err) {
			return b, err
		}
	}
	return backend{}, fmt.Errorf("not found: %s", id)
}
