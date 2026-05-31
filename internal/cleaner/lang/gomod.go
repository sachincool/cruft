package lang

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/sachincool/cruft/internal/cleaner"
	"github.com/sachincool/cruft/internal/fsutil"
)

type goModCleaner struct{}

func init() { cleaner.Register(&goModCleaner{}) }

func (goModCleaner) Name() string               { return "gomod" }
func (goModCleaner) Category() cleaner.Category { return cleaner.CategoryLangPkg }
func (goModCleaner) Description() string {
	return "Go module download cache ($GOMODCACHE). Next `go build` re-downloads all modules — can be slow (5–20 GB common)."
}
func (goModCleaner) Risky() bool             { return false }
func (goModCleaner) RiskReason() string      { return "" }
func (goModCleaner) BusyProcesses() []string { return []string{"go", "gopls"} }

func (goModCleaner) Detect(ctx context.Context) bool {
	if !fsutil.CommandExists("go") {
		return false
	}
	return modCacheDir(ctx) != ""
}

func modCacheDir(ctx context.Context) string {
	out, err := exec.CommandContext(ctx, "go", "env", "GOMODCACHE").Output()
	if err != nil {
		return ""
	}
	dir := strings.TrimSpace(string(out))
	if dir == "" || !fsutil.Exists(dir) {
		return ""
	}
	return dir
}

func (goModCleaner) Scan(ctx context.Context, opts cleaner.ScanOpts) ([]cleaner.Finding, error) {
	dir := modCacheDir(ctx)
	if dir == "" {
		return nil, nil
	}
	s, err := fsutil.Size(ctx, dir, opts.SamplePaths)
	if err != nil || s.Bytes == 0 {
		return nil, err
	}
	samples := make([]cleaner.SamplePath, 0, len(s.SamplePaths))
	for _, sp := range s.SamplePaths {
		samples = append(samples, cleaner.SamplePath{Path: sp.Path, Bytes: sp.Bytes, LastModified: sp.LastModified})
	}
	return []cleaner.Finding{{
		Path:         dir,
		Bytes:        s.Bytes,
		LastModified: s.LastModified,
		Reason:       "go module cache (re-download next build)",
		SamplePaths:  samples,
	}}, nil
}

func (goModCleaner) Execute(ctx context.Context, findings []cleaner.Finding, opts cleaner.ExecOpts) (cleaner.Result, error) {
	res := cleaner.Result{
		Cleaner:  "gomod",
		Findings: len(findings),
		DryRun:   opts.DryRun,
	}
	for _, f := range findings {
		entry := cleaner.AuditEntry{
			Timestamp:  time.Now(),
			RunID:      opts.RunID,
			Cleaner:    "gomod",
			Path:       f.Path,
			Bytes:      f.Bytes,
			DryRun:     opts.DryRun,
			Tombstoned: false,
		}
		if opts.DryRun {
			entry.Success = true
			opts.AuditLog.Record(entry)
			res.Succeeded++
			res.BytesFreed += f.Bytes
			continue
		}
		// Go modcache files are read-only; the cleanest way is `go clean -modcache`.
		cmd := exec.CommandContext(ctx, "go", "clean", "-modcache")
		if err := cmd.Run(); err != nil {
			entry.Error = err.Error()
			opts.AuditLog.Record(entry)
			res.Failed++
			res.Errors = append(res.Errors, err)
			continue
		}
		entry.Success = true
		opts.AuditLog.Record(entry)
		res.Succeeded++
		res.BytesFreed += f.Bytes
	}
	return res, nil
}
