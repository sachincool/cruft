// Package fsutil provides safe filesystem helpers used by cleaners.
//
// The single most important function in the program is SafeRemove: it
// is the chokepoint through which every deletion must pass. It refuses
// to touch any path outside the caller-supplied whitelist, even when
// reached via symlinks.
package fsutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrOutsideWhitelist is returned when a path falls outside the allowed prefixes.
var ErrOutsideWhitelist = errors.New("fsutil: path outside whitelist")

// ErrRefuseRoot is returned when running as root.
var ErrRefuseRoot = errors.New("fsutil: refusing to operate as root")

// SafeRemove deletes path (file or directory tree) iff its
// symlink-resolved location lies under at least one allowed prefix.
//
// The whitelist check uses filepath.EvalSymlinks so that a symlink
// inside an allowed prefix cannot redirect the deletion to an outside
// location. Each allowedPrefix is itself normalised the same way.
//
// If path does not exist, SafeRemove returns nil — already-gone is the
// desired state.
func SafeRemove(path string, allowedPrefixes []string) error {
	if os.Geteuid() == 0 {
		return ErrRefuseRoot
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("fsutil: abs: %w", err)
	}
	// Resolve any symlinks that compose the path. If the path itself
	// doesn't exist, fall back to the lexical abs path — there's
	// nothing to symlink-escape to.
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		// EvalSymlinks fails on broken symlinks; treat as not-exist.
		resolved = abs
	}
	if !underAny(resolved, allowedPrefixes) {
		return fmt.Errorf("%w: %s", ErrOutsideWhitelist, resolved)
	}
	return os.RemoveAll(resolved)
}

// IsUnder returns true if path (after symlink resolution where possible)
// is equal to or beneath any of the allowed prefixes.
func IsUnder(path string, allowedPrefixes []string) (bool, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			resolved = abs
		} else {
			resolved = abs
		}
	}
	return underAny(resolved, allowedPrefixes), nil
}

func underAny(path string, prefixes []string) bool {
	for _, p := range prefixes {
		absP, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		// Resolve the prefix too; a prefix may itself be a symlink
		// (e.g. /var → /private/var on macOS).
		if r, err := filepath.EvalSymlinks(absP); err == nil {
			absP = r
		}
		absP = strings.TrimRight(absP, string(filepath.Separator))
		if path == absP {
			return true
		}
		if strings.HasPrefix(path, absP+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// HomeDir returns the user's home directory or panics on failure
// (a healthy macOS dev laptop always has $HOME).
func HomeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("fsutil: UserHomeDir: %v", err))
	}
	return h
}

// Expand replaces a leading ~ in path with the home directory.
func Expand(path string) string {
	if path == "~" {
		return HomeDir()
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(HomeDir(), path[2:])
	}
	return path
}

// Exists is true if the path exists (file or directory).
func Exists(path string) bool {
	_, err := os.Stat(Expand(path))
	return err == nil
}
