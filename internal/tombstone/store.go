// Package tombstone implements the recovery safety net: instead of
// deleting files outright, cleaners move them under a per-run
// tombstone directory. After a retention window, they're swept.
// Users can restore from any run still within its window.
package tombstone

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	manifestName = "manifest.jsonl"
	dirPerm      = 0o755
)

// Store is rooted at a single base directory (typically
// ~/.local/share/cruft/tombstone) and manages many runs.
type Store struct {
	BaseDir string
}

// New constructs a Store with the given base directory.
func New(baseDir string) *Store { return &Store{BaseDir: baseDir} }

// Bury moves src to the tombstone for runID, preserving its absolute
// path under <BaseDir>/<runID>/files/<original-abs-path>.
//
// Returns the tombstone path on success. Same-filesystem renames are
// atomic; cross-filesystem moves fall back to copy+delete.
func (s *Store) Bury(runID, cleaner, src string) (string, error) {
	if s.BaseDir == "" {
		return "", errors.New("tombstone: empty BaseDir")
	}
	abs, err := filepath.Abs(src)
	if err != nil {
		return "", err
	}
	// Layout: <BaseDir>/<runID>/files/<abs-path-as-relative>
	// e.g. /Users/x/.local/share/cruft/tombstone/2026-…/files/Users/x/.npm
	relAbs := strings.TrimPrefix(abs, string(filepath.Separator))
	dst := filepath.Join(s.BaseDir, runID, "files", relAbs)
	if err := os.MkdirAll(filepath.Dir(dst), dirPerm); err != nil {
		return "", fmt.Errorf("tombstone: mkdir dst: %w", err)
	}
	// Write the manifest entry before moving anything: if the move then
	// fails the entry is merely stale (Restore skips entries whose Src
	// still exists), whereas a manifest failure after the move would
	// strand the data in the tombstone with no record of it.
	if err := s.appendManifest(runID, manifestEntry{
		Buried:  time.Now(),
		Cleaner: cleaner,
		Src:     abs,
		Dst:     dst,
	}); err != nil {
		return "", fmt.Errorf("tombstone: manifest write failed: %w", err)
	}
	if err := os.Rename(abs, dst); err != nil {
		// Cross-device or other; try a recursive copy.
		if cerr := copyTree(abs, dst); cerr != nil {
			return "", fmt.Errorf("tombstone: rename failed (%v); copy fallback failed: %w", err, cerr)
		}
		if cerr := os.RemoveAll(abs); cerr != nil {
			return "", fmt.Errorf("tombstone: copied but couldn't remove src: %w", cerr)
		}
	}
	return dst, nil
}

// Restore moves every file from runID back to its original location.
// Files whose original location is currently occupied are skipped and
// returned in the skipped slice.
func (s *Store) Restore(runID string) (restored, skipped []string, err error) {
	entries, err := s.readManifest(runID)
	if err != nil {
		return nil, nil, err
	}
	for _, e := range entries {
		if _, err := os.Lstat(e.Src); err == nil {
			skipped = append(skipped, e.Src)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(e.Src), dirPerm); err != nil {
			return restored, skipped, fmt.Errorf("tombstone: mkdir for restore: %w", err)
		}
		if err := os.Rename(e.Dst, e.Src); err != nil {
			if cerr := copyTree(e.Dst, e.Src); cerr != nil {
				return restored, skipped, fmt.Errorf("tombstone: restore %s: %w", e.Src, cerr)
			}
			_ = os.RemoveAll(e.Dst)
		}
		restored = append(restored, e.Src)
	}
	// If we restored everything, drop the run dir.
	if len(skipped) == 0 {
		_ = os.RemoveAll(filepath.Join(s.BaseDir, runID))
	}
	return restored, skipped, nil
}

// Sweep deletes runs older than the given retention. Returns the
// run IDs that were swept.
func (s *Store) Sweep(retention time.Duration) ([]string, error) {
	if s.BaseDir == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(s.BaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	cutoff := time.Now().Add(-retention)
	var swept []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			if err := os.RemoveAll(filepath.Join(s.BaseDir, e.Name())); err == nil {
				swept = append(swept, e.Name())
			}
		}
	}
	return swept, nil
}

// Size returns total bytes used by the tombstone directory.
func (s *Store) Size() int64 {
	bytes, _ := dirStats(s.BaseDir)
	return bytes
}

type manifestEntry struct {
	Buried  time.Time `json:"buried"`
	Cleaner string    `json:"cleaner"`
	Src     string    `json:"src"`
	Dst     string    `json:"dst"`
}

func (s *Store) manifestPath(runID string) string {
	return filepath.Join(s.BaseDir, runID, manifestName)
}

func (s *Store) appendManifest(runID string, e manifestEntry) error {
	mp := s.manifestPath(runID)
	if err := os.MkdirAll(filepath.Dir(mp), dirPerm); err != nil {
		return err
	}
	f, err := os.OpenFile(mp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	return enc.Encode(e)
}

func (s *Store) readManifest(runID string) ([]manifestEntry, error) {
	f, err := os.Open(s.manifestPath(runID))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []manifestEntry
	dec := json.NewDecoder(f)
	for dec.More() {
		var e manifestEntry
		if err := dec.Decode(&e); err != nil {
			return out, err
		}
		out = append(out, e)
	}
	return out, nil
}

func copyTree(src, dst string) error {
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if !srcInfo.IsDir() {
		return copyFile(src, dst, srcInfo)
	}
	if err := os.MkdirAll(dst, srcInfo.Mode().Perm()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		s := filepath.Join(src, e.Name())
		d := filepath.Join(dst, e.Name())
		if err := copyTree(s, d); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(target, dst)
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	// Close errors matter here: a short write surfacing at close time
	// means the tombstone copy is incomplete and the source must not
	// be deleted.
	return out.Close()
}

func dirStats(path string) (int64, int) {
	var bytes int64
	var count int
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return nil
		}
		if d.Type().IsRegular() {
			if fi, err := d.Info(); err == nil {
				bytes += fi.Size()
				count++
			}
		}
		return nil
	})
	return bytes, count
}
