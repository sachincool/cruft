package cleaner

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/sachincool/cruft/internal/fsutil"
	"github.com/sachincool/cruft/internal/tombstone"
)

// PathCleaner is a reusable building block for cleaners that just need
// to delete (or tombstone) a fixed set of directories under $HOME.
//
// Concrete cleaners embed it (or just construct one) and rely on the
// shared Scan/Execute implementation. ShellOut cleaners (brew, docker)
// roll their own Execute and don't use this.
type PathCleaner struct {
	NameValue     string
	CategoryValue Category
	Desc          string
	IsRisky       bool
	// RiskReasonValue is the short human-readable explanation of *why*
	// this cleaner is risky. Required when IsRisky=true; ignored otherwise.
	// Examples: "may corrupt VM if running", "forces multi-minute re-index".
	RiskReasonValue string
	// Paths is the set of absolute (or ~-prefixed) paths this cleaner
	// owns and is allowed to remove. Doubles as the SafeRemove whitelist.
	Paths []string
	// DetectCmd, if non-empty, requires this command to be on $PATH
	// before Detect returns true.
	DetectCmd string
	// DetectAnyPath: if true, Detect returns true when ANY of Paths
	// exists (default false → all must exist).
	DetectAnyPath bool
	// BusyProcs are process names to check via pgrep.
	BusyProcs []string
	// Reason is the human-readable Finding.Reason.
	Reason string
}

func (p *PathCleaner) Name() string            { return p.NameValue }
func (p *PathCleaner) Category() Category      { return p.CategoryValue }
func (p *PathCleaner) Description() string     { return p.Desc }
func (p *PathCleaner) Risky() bool             { return p.IsRisky }
func (p *PathCleaner) RiskReason() string      { return p.RiskReasonValue }
func (p *PathCleaner) BusyProcesses() []string { return p.BusyProcs }

func (p *PathCleaner) Detect(ctx context.Context) bool {
	if p.DetectCmd != "" && !fsutil.CommandExists(p.DetectCmd) {
		return false
	}
	if len(p.Paths) == 0 {
		return false
	}
	if p.DetectAnyPath {
		for _, path := range p.Paths {
			if fsutil.Exists(path) {
				return true
			}
		}
		return false
	}
	for _, path := range p.Paths {
		if !fsutil.Exists(path) {
			return false
		}
	}
	return true
}

// Scan returns one Finding per path that exists and is non-empty.
func (p *PathCleaner) Scan(ctx context.Context, opts ScanOpts) ([]Finding, error) {
	var out []Finding
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, path := range p.Paths {
		path := fsutil.Expand(path)
		if !fsutil.Exists(path) {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			s, err := fsutil.Size(ctx, path, opts.SamplePaths)
			if err != nil || s.Bytes == 0 {
				return
			}
			sp := make([]SamplePath, 0, len(s.SamplePaths))
			for _, x := range s.SamplePaths {
				sp = append(sp, SamplePath{Path: x.Path, Bytes: x.Bytes, LastModified: x.LastModified})
			}
			f := Finding{
				Path:         path,
				Bytes:        s.Bytes,
				LastModified: s.LastModified,
				Reason:       p.Reason,
				Risky:        p.IsRisky,
				SamplePaths:  sp,
			}
			mu.Lock()
			out = append(out, f)
			mu.Unlock()
		}()
	}
	wg.Wait()
	return out, nil
}

// Execute deletes (or tombstones) each approved finding's Path.
func (p *PathCleaner) Execute(ctx context.Context, findings []Finding, opts ExecOpts) (Result, error) {
	res := Result{
		Cleaner:  p.NameValue,
		Findings: len(findings),
		DryRun:   opts.DryRun,
	}
	whitelist := make([]string, 0, len(p.Paths))
	for _, x := range p.Paths {
		whitelist = append(whitelist, fsutil.Expand(x))
	}
	for _, f := range findings {
		if err := ctx.Err(); err != nil {
			break
		}
		entry := AuditEntry{
			Timestamp:  time.Now(),
			RunID:      opts.RunID,
			Cleaner:    p.NameValue,
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
		// Confirm path is still under whitelist (defence-in-depth — the
		// cleaner's Paths can't have changed at runtime, but a malicious
		// symlink could).
		ok, err := fsutil.IsUnder(f.Path, whitelist)
		if err != nil || !ok {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("refused: %s outside whitelist", f.Path))
			entry.Error = "outside whitelist"
			opts.AuditLog.Record(entry)
			continue
		}
		if opts.UseTombstone && opts.Tombstone != "" {
			// Move into tombstone instead of deleting.
			store := tombstone.New(filepath.Dir(opts.Tombstone))
			runID := filepath.Base(opts.Tombstone)
			if _, err := store.Bury(runID, p.NameValue, f.Path); err != nil {
				res.Failed++
				res.Errors = append(res.Errors, err)
				entry.Error = err.Error()
				opts.AuditLog.Record(entry)
				continue
			}
		} else {
			if err := fsutil.SafeRemove(f.Path, whitelist); err != nil {
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
