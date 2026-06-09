package fsutil

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StaleDir is a directory matching some criterion ("named .terraform/")
// whose parent project has not been touched recently.
type StaleDir struct {
	Path             string // absolute path to the matching directory
	ProjectDir       string // parent project root
	NewestProjectMod time.Time
	AgeDays          int
}

// FindStaleDirs walks each search root, finds every directory whose
// name matches dirName, and returns those whose parent project (the
// directory containing the match) has no project-marker file newer
// than staleDays.
//
// projectMarkerExts is the set of file extensions whose modtime defines
// "project activity" (e.g. [".tf"] for terraform, [".hcl"] for terragrunt).
// If empty, every regular file in the project counts.
//
// Hidden dirs other than dirName itself are skipped, as are common
// vendor/cache directories.
func FindStaleDirs(
	ctx context.Context,
	searchRoots []string,
	dirName string,
	projectMarkerExts []string,
	staleDays int,
) ([]StaleDir, error) {
	threshold := time.Now().AddDate(0, 0, -staleDays)
	var out []StaleDir

	skipDirs := map[string]bool{
		"node_modules": true,
		".git":         true,
		"vendor":       true,
		"target":       true,
		"build":        true,
		"dist":         true,
		"Pods":         true,
	}

	for _, root := range searchRoots {
		root = Expand(root)
		if !Exists(root) {
			continue
		}
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, werr error) error {
			if werr != nil {
				return nil
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if !d.IsDir() {
				return nil
			}
			name := d.Name()
			if name == dirName {
				project := filepath.Dir(p)
				// The marker gates "is this really a <kind> project";
				// without one we refuse to touch the directory at all.
				marker := newestMarker(project, projectMarkerExts, dirName)
				if marker.IsZero() {
					return filepath.SkipDir
				}
				// Staleness must reflect real activity anywhere in the
				// project tree, not just the root markers: a project edited
				// daily under src/ can have a package.json untouched for
				// months, and deleting its node_modules would hit an
				// actively developed project.
				newest := newestActivity(ctx, project, dirName, threshold, marker)
				if newest.Before(threshold) {
					age := int(time.Since(newest).Hours() / 24)
					out = append(out, StaleDir{
						Path:             p,
						ProjectDir:       project,
						NewestProjectMod: newest,
						AgeDays:          age,
					})
				}
				// Don't descend into the matched dir; nothing useful there.
				return filepath.SkipDir
			}
			// Skip common bulky / irrelevant dirs.
			if skipDirs[name] {
				return filepath.SkipDir
			}
			// Skip other dotdirs.
			if strings.HasPrefix(name, ".") && p != root {
				return filepath.SkipDir
			}
			return nil
		})
		if err != nil {
			return out, err
		}
	}
	return out, nil
}

// newestActivity returns the newest modtime of any regular file in the
// project tree, excluding the artifact dir being judged, other build/
// vendor output, and hidden dirs. It short-circuits as soon as it sees
// a file newer than threshold — for active projects (the common case
// that must NOT be deleted) this exits on the first recent file, so the
// walk stays cheap. Starts from seed (the marker mtime) so the result
// is never older than what the markers already prove.
func newestActivity(ctx context.Context, projectDir, ignoreDir string, threshold, seed time.Time) time.Time {
	newest := seed
	skipDirs := map[string]bool{
		"node_modules": true, ".git": true, "vendor": true,
		"target": true, "build": true, "dist": true, "Pods": true,
	}
	_ = filepath.WalkDir(projectDir, func(p string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if p == projectDir {
				return nil
			}
			if name == ignoreDir || skipDirs[name] || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return nil
		}
		if fi.ModTime().After(newest) {
			newest = fi.ModTime()
			if newest.After(threshold) {
				return filepath.SkipAll // project is active; stop walking
			}
		}
		return nil
	})
	return newest
}

// newestMarker returns the latest modtime among project-marker files
// directly under projectDir (non-recursive). The matched dirName itself
// is ignored. Returns zero time if no markers found.
func newestMarker(projectDir string, exts []string, ignoreDir string) time.Time {
	var newest time.Time
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return newest
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == ignoreDir {
			continue
		}
		if len(exts) > 0 {
			match := false
			for _, ext := range exts {
				if strings.HasSuffix(e.Name(), ext) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		if fi, err := e.Info(); err == nil {
			if fi.ModTime().After(newest) {
				newest = fi.ModTime()
			}
		}
	}
	return newest
}
