package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/git-context/internal/model"
)

var lockCmd = &cobra.Command{
	Use:   "lock <target>",
	Short: "Lock a task or file path",
	Long: `Lock a target to prevent conflicts.

Targets can be task IDs or file paths.

Examples:
  git ctx lock task-abc123      # Lock a task
  git ctx lock src/auth/        # Lock a directory
  git ctx lock --shared task-1  # Shared lock (syncs)`,
	Args: cobra.ExactArgs(1),
	RunE: runLock,
}

var lockListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all locks",
	RunE:  runLockList,
}

var unlockCmd = &cobra.Command{
	Use:   "unlock [target]",
	Short: "Release lock(s)",
	Long: `Release locks you own.

Without arguments, releases all your locks.

Examples:
  git ctx unlock task-abc123   # Unlock specific target
  git ctx unlock               # Unlock all your locks`,
	RunE: runUnlock,
}

func init() {
	lockCmd.AddCommand(lockListCmd)
	rootCmd.AddCommand(unlockCmd)
}

func runLock(cmd *cobra.Command, args []string) error {
	target := args[0]
	author := model.GetAuthorShort()
	
	// Check if already locked (either storage)
	existingLocal, _ := store.Local.ReadLock(target)
	existingShared, _ := store.Shared.ReadLock(target)
	
	if existingLocal != nil && !existingLocal.IsExpired() {
		return fmt.Errorf("already locked by %s (expires: %s)", 
			existingLocal.LockedBy, existingLocal.ExpiresAt.Format("15:04"))
	}
	if existingShared != nil && !existingShared.IsExpired() {
		return fmt.Errorf("already locked by %s (expires: %s)", 
			existingShared.LockedBy, existingShared.ExpiresAt.Format("15:04"))
	}
	
	// Create lock
	lock := model.NewLock(target, author)
	
	storage := getStorage()
	if err := storage.WriteLock(lock); err != nil {
		return fmt.Errorf("failed to lock: %w", err)
	}
	
	storageType := "local"
	if flagShared {
		storageType = "shared"
	}
	fmt.Printf("Locked (%s): %s\n", storageType, target)
	
	return nil
}

func runLockList(cmd *cobra.Command, args []string) error {
	var locks []*model.Lock
	
	// Collect locks based on flags
	if flagAll || !flagShared {
		local, err := store.Local.ListLocks()
		if err == nil {
			locks = append(locks, local...)
		}
	}
	
	if flagAll || flagShared {
		shared, err := store.Shared.ListLocks()
		if err == nil {
			locks = append(locks, shared...)
		}
	}
	
	if len(locks) == 0 {
		fmt.Println("No active locks")
		return nil
	}
	
	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	
	fmt.Fprintln(w, "TARGET\tLOCKED BY\tEXPIRES")
	fmt.Fprintln(w, "------\t---------\t-------")
	
	for _, l := range locks {
		if !l.IsExpired() {
			fmt.Fprintf(w, "%s\t%s\t%s\n", l.Target, l.LockedBy, l.ExpiresAt.Format("15:04"))
		}
	}
	
	return w.Flush()
}

func runUnlock(cmd *cobra.Command, args []string) error {
	author := model.GetAuthorShort()
	
	if len(args) == 0 {
		// Unlock all owned by current user
		return unlockAll(author)
	}
	
	target := args[0]
	
	// Try to find and unlock
	localLock, _ := store.Local.ReadLock(target)
	if localLock != nil {
		if !localLock.IsOwnedBy(author) {
			return fmt.Errorf("cannot unlock: owned by %s", localLock.LockedBy)
		}
		store.Local.DeleteLock(target)
		fmt.Printf("Unlocked (local): %s\n", target)
		return nil
	}
	
	sharedLock, _ := store.Shared.ReadLock(target)
	if sharedLock != nil {
		if !sharedLock.IsOwnedBy(author) {
			return fmt.Errorf("cannot unlock: owned by %s", sharedLock.LockedBy)
		}
		store.Shared.DeleteLock(target)
		fmt.Printf("Unlocked (shared): %s\n", target)
		return nil
	}
	
	return fmt.Errorf("not locked: %s", target)
}

func unlockAll(author string) error {
	count := 0
	
	// Local locks
	localLocks, _ := store.Local.ListLocks()
	for _, l := range localLocks {
		if l.IsOwnedBy(author) {
			store.Local.DeleteLock(l.Target)
			fmt.Printf("Unlocked (local): %s\n", l.Target)
			count++
		}
	}
	
	// Shared locks
	sharedLocks, _ := store.Shared.ListLocks()
	for _, l := range sharedLocks {
		if l.IsOwnedBy(author) {
			store.Shared.DeleteLock(l.Target)
			fmt.Printf("Unlocked (shared): %s\n", l.Target)
			count++
		}
	}
	
	if count == 0 {
		fmt.Println("No locks to release")
	}
	
	return nil
}

