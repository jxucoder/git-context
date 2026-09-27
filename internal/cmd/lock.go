package cmd

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/jxucoder/git-context/internal/model"
	"github.com/jxucoder/git-context/internal/storage"
	"github.com/spf13/cobra"
)

var lockCmd = &cobra.Command{
	Use:   "lock <target>",
	Short: "Lock a task or file path",
	Long: `Lock a target to prevent conflicts.

Targets can be task IDs or file paths. Locks expire after 4 hours, and
locking a target you already hold refreshes the expiry.

Examples:
  git ctx lock task-abc123      # Lock a task
  git ctx lock src/auth/        # Lock a directory`,
	Args: cobra.ExactArgs(1),
	RunE: runLock,
}

var lockListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active locks",
	RunE:  runLockList,
}

var unlockCmd = &cobra.Command{
	Use:   "unlock [target]",
	Short: "Release lock(s)",
	Long: `Release locks you own. Expired locks can be released by anyone.

Without arguments, releases all your locks and clears any expired locks.

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
	target := model.NormalizeTarget(args[0])
	if target == "" {
		return fmt.Errorf("lock target must not be empty")
	}
	author := model.GetAuthorShort()
	b := targetBackend()

	// A live lock in any other backend blocks too.
	for _, other := range allBackends() {
		if other.name == b.name {
			continue
		}
		existing, err := other.ReadLock(target)
		if err != nil {
			if unavailable(err) || errors.Is(err, storage.ErrCorrupt) {
				continue
			}
			return err
		}
		if !existing.IsExpired() && !existing.IsOwnedBy(author) {
			return lockedError(existing)
		}
	}

	if err := b.WriteLock(model.NewLock(target, author)); err != nil {
		var locked *storage.LockedError
		if errors.As(err, &locked) {
			return lockedError(locked.Lock)
		}
		return fmt.Errorf("failed to lock: %w", err)
	}

	fmt.Printf("Locked (%s): %s\n", b.name, target)
	return nil
}

func lockedError(l *model.Lock) error {
	return fmt.Errorf("already locked by %s (expires %s)", l.LockedBy, formatExpiry(l.ExpiresAt))
}

// formatExpiry shows a lock's expiry in the local time zone.
func formatExpiry(t time.Time) string {
	return t.Local().Format("2006-01-02 15:04 MST")
}

func runLockList(cmd *cobra.Command, args []string) error {
	locks, err := collect(selectedBackends(), backend.ListLocks, nil)
	if err != nil {
		return err
	}

	locks = slices.DeleteFunc(locks, (*model.Lock).IsExpired)
	slices.SortFunc(locks, func(a, b *model.Lock) int { return strings.Compare(a.Target, b.Target) })

	if flagJSON {
		return printJSON(locks)
	}

	if len(locks) == 0 {
		fmt.Println("No active locks")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "TARGET\tLOCKED BY\tEXPIRES")
	fmt.Fprintln(w, "------\t---------\t-------")

	for _, l := range locks {
		fmt.Fprintf(w, "%s\t%s\t%s\n", l.Target, l.LockedBy, formatExpiry(l.ExpiresAt))
	}

	return w.Flush()
}

func runUnlock(cmd *cobra.Command, args []string) error {
	author := model.GetAuthorShort()

	if len(args) == 0 {
		return unlockAll(author)
	}

	target := model.NormalizeTarget(args[0])

	for _, b := range allBackends() {
		l, err := b.ReadLock(target)
		switch {
		case unavailable(err):
			continue
		case errors.Is(err, storage.ErrCorrupt):
			// An unreadable lock file holds nothing useful; anyone may clear it.
		case err != nil:
			return err
		case !l.IsExpired() && !l.IsOwnedBy(author):
			return fmt.Errorf("cannot unlock: owned by %s (expires %s)", l.LockedBy, formatExpiry(l.ExpiresAt))
		}

		if err := b.DeleteLock(target); err != nil && !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("failed to unlock: %w", err)
		}
		fmt.Printf("Unlocked (%s): %s\n", b.name, target)
		return nil
	}

	return fmt.Errorf("not locked: %s", target)
}

// unlockAll releases every lock the author holds and clears expired locks
// left behind by anyone.
func unlockAll(author string) error {
	count := 0

	for _, b := range allBackends() {
		locks, err := b.ListLocks()
		if unavailable(err) {
			continue
		}
		if err := warnSkipped(err); err != nil {
			return err
		}

		for _, l := range locks {
			var action string
			switch {
			case l.IsOwnedBy(author):
				action = fmt.Sprintf("Unlocked (%s): %s", b.name, l.Target)
			case l.IsExpired():
				action = fmt.Sprintf("Removed expired lock (%s): %s held by %s", b.name, l.Target, l.LockedBy)
			default:
				continue
			}
			if err := b.DeleteLock(l.Target); err != nil {
				warnf("failed to unlock %s: %v", l.Target, err)
				continue
			}
			fmt.Println(action)
			count++
		}
	}

	if count == 0 {
		fmt.Println("No locks to release")
	}

	return nil
}
