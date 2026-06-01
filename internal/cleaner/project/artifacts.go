// Package project holds cleaners that reclaim rebuildable build artifacts
// (node_modules/, target/, build/, …) sitting inside project trees the
// user hasn't touched in --stale-days. Unlike the lang/system cache
// cleaners, these walk your project roots rather than fixed cache dirs.
package project

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/sachincool/cruft/internal/cleaner"
	"github.com/sachincool/cruft/internal/fsutil"
	"github.com/sachincool/cruft/internal/tombstone"
)

type artifactsCleaner struct{}

func init() { cleaner.Register(&artifactsCleaner{}) }

func (artifactsCleaner) Name() string                    { return "project-artifacts" }
func (artifactsCleaner) Category() cleaner.Category      { return cleaner.CategoryProject }
func (artifactsCleaner) Risky() bool                     { return false }
func (artifactsCleaner) RiskReason() string              { return "" }
func (artifactsCleaner) BusyProcesses() []string         { return nil }
func (artifactsCleaner) Detect(ctx context.Context) bool { return true }
func (artifactsCleaner) Description() string {
	return "Stale build artifacts (node_modules/, target/, build/, .build/, vendor/, dist/) under projects you haven't touched in --stale-days. Rebuilds with the project's install/build command."
}

// artifactKind pairs a build-output directory name with the
// project-marker files (matched by filename suffix in newestMarker)
// that signal the surrounding project's source activity.
type artifactKind struct {
	dir     string   // directory to reclaim
	markers []string // marker filenames defining "project touched"
	label   string   // human project type, for the finding reason
}

// kinds covers cleardisk's project types: Node.js, Rust, Maven, Swift,
// Gradle, CMake, Flutter, Go, PHP, Ruby, and JS bundles.
var kinds = []artifactKind{
	{"node_modules", []string{"package.json"}, "Node.js"},
	{"target", []string{"Cargo.toml", "pom.xml"}, "Rust/Maven"},
	{".build", []string{"Package.swift"}, "Swift"},
	{"build", []string{"build.gradle", "build.gradle.kts", "CMakeLists.txt", "pubspec.yaml"}, "Gradle/CMake/Flutter"},
	{"vendor", []string{"go.mod", "composer.json", "Gemfile"}, "Go/PHP/Ruby"},
	{"dist", []string{"package.json"}, "JS bundle"},
}

func (artifactsCleaner) Scan(ctx context.Context, opts cleaner.ScanOpts) ([]cleaner.Finding, error) {
	roots := opts.SearchRoots
	if len(roots) == 0 {
		roots = defaultSearchRoots()
	}
	var out []cleaner.Finding
	for _, k := range kinds {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		stale, err := fsutil.FindStaleDirs(ctx, roots, k.dir, k.markers, opts.StaleDays)
		if err != nil {
			return out, err
		}
		for _, d := range stale {
			s, err := fsutil.Size(ctx, d.Path, opts.SamplePaths)
			if err != nil || s.Bytes == 0 {
				continue
			}
			samples := make([]cleaner.SamplePath, 0, len(s.SamplePaths))
			for _, sp := range s.SamplePaths {
				samples = append(samples, cleaner.SamplePath{Path: sp.Path, Bytes: sp.Bytes, LastModified: sp.LastModified})
			}
			out = append(out, cleaner.Finding{
				Path:         d.Path,
				Bytes:        s.Bytes,
				LastModified: d.NewestProjectMod,
				Reason:       fmt.Sprintf("%s %s/ — project untouched %dd in %s", k.label, k.dir, d.AgeDays, d.ProjectDir),
				SamplePaths:  samples,
			})
		}
	}
	return out, nil
}

func (artifactsCleaner) Execute(ctx context.Context, findings []cleaner.Finding, opts cleaner.ExecOpts) (cleaner.Result, error) {
	const name = "project-artifacts"
	res := cleaner.Result{Cleaner: name, Findings: len(findings), DryRun: opts.DryRun}
	for _, f := range findings {
		if err := ctx.Err(); err != nil {
			break
		}
		entry := cleaner.AuditEntry{
			Timestamp:  time.Now(),
			RunID:      opts.RunID,
			Cleaner:    name,
			Path:       f.Path,
			Bytes:      f.Bytes,
			DryRun:     opts.DryRun,
			Tombstoned: opts.UseTombstone,
		}
		if opts.DryRun {
			entry.Success = true
			opts.AuditLog.Record(entry)
			res.Succeeded++
			res.BytesFreed += f.Bytes
			continue
		}
		if opts.UseTombstone && opts.Tombstone != "" {
			store := tombstone.New(filepath.Dir(opts.Tombstone))
			if _, err := store.Bury(filepath.Base(opts.Tombstone), name, f.Path); err != nil {
				res.Failed++
				res.Errors = append(res.Errors, err)
				entry.Error = err.Error()
				opts.AuditLog.Record(entry)
				continue
			}
		} else {
			// Whitelist scoped to exactly this finding's path.
			if err := fsutil.SafeRemove(f.Path, []string{f.Path}); err != nil {
				res.Failed++
				res.Errors = append(res.Errors, err)
				entry.Error = err.Error()
				opts.AuditLog.Record(entry)
				continue
			}
		}
		entry.Success = true
		opts.AuditLog.Record(entry)
		res.Succeeded++
		res.BytesFreed += f.Bytes
	}
	return res, nil
}

// defaultSearchRoots mirrors the IaC cleaners' default project locations.
func defaultSearchRoots() []string {
	return []string{"~/Projects", "~/projects", "~/code", "~/work", "~/src", "~/dev"}
}
