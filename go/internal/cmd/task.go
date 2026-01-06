package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/git-context/internal/model"
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
  git ctx task add "Setup database" -d "PostgreSQL schema with users table"
  git ctx task add --shared "Team task"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runTaskAdd,
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long: `List tasks from storage.

Examples:
  git ctx task list           # Local tasks
  git ctx task list --shared  # Shared tasks
  git ctx task list --all     # Everything`,
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
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskClaim,
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
	author := model.GetAuthorShort()
	
	t := model.NewTask(title, taskDescription, author, flagShared)
	
	storage := getStorage()
	if err := storage.WriteTask(t); err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	
	storageType := "local"
	if flagShared {
		storageType = "shared"
	}
	fmt.Printf("Created (%s): %s\n", storageType, t.ID)
	
	return nil
}

func runTaskList(cmd *cobra.Command, args []string) error {
	var tasks []*model.Task
	
	// Collect tasks based on flags
	if flagAll || !flagShared {
		local, err := store.Local.ListTasks()
		if err != nil {
			return fmt.Errorf("failed to list local tasks: %w", err)
		}
		for _, t := range local {
			t.Shared = false
			tasks = append(tasks, t)
		}
	}
	
	if flagAll || flagShared {
		shared, err := store.Shared.ListTasks()
		if err != nil {
			return fmt.Errorf("failed to list shared tasks: %w", err)
		}
		for _, t := range shared {
			t.Shared = true
			tasks = append(tasks, t)
		}
	}
	
	if flagJSON {
		data, err := json.MarshalIndent(tasks, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
	
	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	
	fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tTYPE\tOWNER")
	fmt.Fprintln(w, "----\t-----\t------\t----\t-----")
	
	for _, t := range tasks {
		typeStr := "[local]"
		if t.Shared {
			typeStr = "[shared]"
		}
		
		title := t.Title
		if len(title) > 35 {
			title = title[:32] + "..."
		}
		
		owner := t.Owner
		if owner == "" {
			owner = "-"
		}
		
		fmt.Fprintf(w, "%s\t%s\t[%s]\t%s\t%s\n", t.ID, title, t.Status, typeStr, owner)
	}
	
	return w.Flush()
}

func runTaskShow(cmd *cobra.Command, args []string) error {
	id := args[0]
	
	t, storageType := findTask(id)
	if t == nil {
		return fmt.Errorf("not found: %s", id)
	}
	
	if flagJSON {
		data, err := json.MarshalIndent(t, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
	
	// Pretty print
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Printf("  %s\n", t.Title)
	fmt.Printf("  Status: %s • Type: %s • Created by: %s\n", t.Status, storageType, t.CreatedBy)
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
	
	t, storageType := findTask(id)
	if t == nil {
		return fmt.Errorf("not found: %s", id)
	}
	
	if t.Status == model.TaskClaimed && t.Owner != author {
		return fmt.Errorf("already claimed by %s", t.Owner)
	}
	
	t.Claim(author)
	
	// Save to correct storage
	var err error
	if storageType == "local" {
		err = store.Local.WriteTask(t)
	} else {
		err = store.Shared.WriteTask(t)
	}
	
	if err != nil {
		return fmt.Errorf("failed to claim: %w", err)
	}
	
	fmt.Printf("Claimed: %s\n", id)
	return nil
}

func runTaskDrop(cmd *cobra.Command, args []string) error {
	id := args[0]
	author := model.GetAuthorShort()
	
	t, storageType := findTask(id)
	if t == nil {
		return fmt.Errorf("not found: %s", id)
	}
	
	if t.Owner != author {
		return fmt.Errorf("not owned by you (owner: %s)", t.Owner)
	}
	
	t.Drop()
	
	// Save to correct storage
	var err error
	if storageType == "local" {
		err = store.Local.WriteTask(t)
	} else {
		err = store.Shared.WriteTask(t)
	}
	
	if err != nil {
		return fmt.Errorf("failed to drop: %w", err)
	}
	
	fmt.Printf("Dropped: %s\n", id)
	return nil
}

func runTaskDone(cmd *cobra.Command, args []string) error {
	id := args[0]
	
	t, storageType := findTask(id)
	if t == nil {
		return fmt.Errorf("not found: %s", id)
	}
	
	t.Done()
	
	// Save to correct storage
	var err error
	if storageType == "local" {
		err = store.Local.WriteTask(t)
	} else {
		err = store.Shared.WriteTask(t)
	}
	
	if err != nil {
		return fmt.Errorf("failed to mark done: %w", err)
	}
	
	fmt.Printf("Done: %s\n", id)
	return nil
}

func runTaskComment(cmd *cobra.Command, args []string) error {
	id := args[0]
	message := args[1]
	author := model.GetAuthorShort()
	
	t, storageType := findTask(id)
	if t == nil {
		return fmt.Errorf("not found: %s", id)
	}
	
	t.AddComment(author, message)
	
	// Save to correct storage
	var err error
	if storageType == "local" {
		err = store.Local.WriteTask(t)
	} else {
		err = store.Shared.WriteTask(t)
	}
	
	if err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}
	
	fmt.Printf("Comment added to: %s\n", id)
	return nil
}

func findTask(id string) (*model.Task, string) {
	// Try local
	t, err := store.Local.ReadTask(id)
	if err == nil && t != nil {
		return t, "local"
	}
	
	// Try shared
	t, err = store.Shared.ReadTask(id)
	if err == nil && t != nil {
		return t, "shared"
	}
	
	return nil, ""
}

