package fsutil

import (
	"context"
	"os/exec"
)

// AnyProcessRunning returns the first name in names for which pgrep
// finds a live process, or "" if none are running.
//
// pgrep is part of macOS base; if it's missing for some reason, this
// degrades to "no process running" (returns "") — the runner will then
// proceed, which matches the user's intent on a non-standard system.
func AnyProcessRunning(ctx context.Context, names []string) string {
	for _, n := range names {
		if isRunning(ctx, n) {
			return n
		}
	}
	return ""
}

func isRunning(ctx context.Context, name string) bool {
	// -x matches the full command name; safer than substring.
	cmd := exec.CommandContext(ctx, "pgrep", "-x", name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// CommandExists returns true if name is on $PATH.
func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
