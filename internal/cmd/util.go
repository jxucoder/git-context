package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jxucoder/git-context/internal/storage"
)

// backend is a storage implementation together with the label shown in output.
type backend struct {
	name   string
	shared bool
	storage.Storage
}

func localBackend() backend {
	return backend{name: "local", Storage: store.Local}
}

func sharedBackend() backend {
	return backend{name: "shared", shared: true, Storage: store.Shared}
}

// allBackends lists every backend, local first, for commands that look an id
// up wherever it is stored.
func allBackends() []backend {
	return []backend{localBackend(), sharedBackend()}
}

// selectedBackends applies --shared and --all; the default is local only.
func selectedBackends() []backend {
	switch {
	case flagAll:
		return allBackends()
	case flagShared:
		return []backend{sharedBackend()}
	default:
		return []backend{localBackend()}
	}
}

// targetBackend is where new entries are written.
func targetBackend() backend {
	if flagShared {
		return sharedBackend()
	}
	return localBackend()
}

// unavailable reports errors that mean "nothing stored here", so a lookup can
// move on to the next backend.
func unavailable(err error) bool {
	return errors.Is(err, storage.ErrNotFound) || errors.Is(err, storage.ErrNotImplemented)
}

// collect merges the items every backend returns. Unavailable backends are
// skipped, entries that could not be read are reported as warnings, and mark
// (if set) tags each item with the backend it came from.
func collect[T any](backends []backend, list func(backend) ([]T, error), mark func(T, backend)) ([]T, error) {
	items := []T{}
	for _, b := range backends {
		found, err := list(b)
		if unavailable(err) {
			continue
		}
		if err := warnSkipped(err); err != nil {
			return nil, fmt.Errorf("failed to list %s: %w", b.name, err)
		}
		for _, item := range found {
			if mark != nil {
				mark(item, b)
			}
			items = append(items, item)
		}
	}
	return items, nil
}

// warnf prints a warning to stderr.
func warnf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "warn: "+format+"\n", args...)
}

// warnSkipped prints the corrupt-entry report a List call returned, one line
// per entry, and returns nil. Any other error is returned unchanged.
func warnSkipped(err error) error {
	if err == nil {
		return nil
	}
	if !errors.Is(err, storage.ErrCorrupt) {
		return err
	}
	for _, line := range strings.Split(err.Error(), "\n") {
		warnf("skipping %s", line)
	}
	return nil
}

// printJSON writes v as indented JSON to stdout.
func printJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// truncate shortens s to at most max characters, ending with "..." when cut.
func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-3]) + "..."
}

// typeLabel is the storage column shown in tables.
func typeLabel(shared bool) string {
	if shared {
		return "[shared]"
	}
	return "[local]"
}
