package iac

import (
	"context"
	"fmt"

	"github.com/sachincool/cruft/internal/cleaner"
	"github.com/sachincool/cruft/internal/fsutil"
)

type tgCleaner struct{}

func init() { cleaner.Register(&tgCleaner{}) }

func (tgCleaner) Name() string               { return "terragrunt" }
func (tgCleaner) Category() cleaner.Category { return cleaner.CategoryIaC }
func (tgCleaner) Description() string {
	return "Stale .terragrunt-cache/ directories under projects you haven't touched in --stale-days."
}
func (tgCleaner) Risky() bool                     { return false }
func (tgCleaner) RiskReason() string              { return "" }
func (tgCleaner) BusyProcesses() []string         { return []string{"terragrunt", "terraform"} }
func (tgCleaner) Detect(ctx context.Context) bool { return true }

func (tgCleaner) Scan(ctx context.Context, opts cleaner.ScanOpts) ([]cleaner.Finding, error) {
	roots := opts.SearchRoots
	if len(roots) == 0 {
		roots = defaultSearchRoots()
	}
	stale, err := fsutil.FindStaleDirs(ctx, roots, ".terragrunt-cache", []string{".hcl"}, opts.StaleDays)
	if err != nil {
		return nil, err
	}
	out := make([]cleaner.Finding, 0, len(stale))
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
			Reason:       fmt.Sprintf("project untouched %dd in %s", d.AgeDays, d.ProjectDir),
			SamplePaths:  samples,
		})
	}
	return out, nil
}

func (tgCleaner) Execute(ctx context.Context, findings []cleaner.Finding, opts cleaner.ExecOpts) (cleaner.Result, error) {
	return executeStale(ctx, "terragrunt", ".terragrunt-cache", findings, opts)
}
