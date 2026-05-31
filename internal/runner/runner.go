// Package runner orchestrates the scan and execute phases.
//
// Phase order: Detect → BusyProcesses check → Scan (parallel) →
// (caller-driven Approve filter) → Execute (parallel) → Result.
package runner

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/sachincool/cruft/internal/audit"
	"github.com/sachincool/cruft/internal/cleaner"
	"github.com/sachincool/cruft/internal/fsutil"
	"github.com/sachincool/cruft/internal/tombstone"
)

// ScanResult is what one cleaner produced during the scan phase.
type ScanResult struct {
	Cleaner      cleaner.Cleaner
	Findings     []cleaner.Finding
	TotalBytes   int64
	Err          error
	NotInstalled bool
	BusyProcess  string // non-empty → skipped because this process is running
	DurationMS   int64
}

// ExecResult adds the execute outcome to a ScanResult.
type ExecResult struct {
	ScanResult
	Result cleaner.Result
}

// Options configures a Runner.
type Options struct {
	MaxParallel  int
	StaleDays    int
	SearchRoots  []string
	SamplePaths  int
	DryRun       bool
	UseTombstone bool
	TombstoneDir string
	AuditDir     string
	Budget       int64 // bytes; 0 = unlimited
	// AutoApproveRisky controls whether findings from risky cleaners
	// are pre-approved for execute. Default false — risky cleaners are
	// still scanned (so they're visible in TUI/summary) but their
	// findings stay unapproved until the user toggles them in the TUI
	// or re-runs with --include-risky.
	AutoApproveRisky bool
}

// Runner is the orchestrator. One instance per `cruft` invocation.
type Runner struct {
	opts      Options
	runID     string
	auditLog  *audit.Log
	tombstone *tombstone.Store
	budget    *Budget
}

// New constructs a runner. opts may be the zero value; sensible
// defaults are applied.
func New(opts Options) (*Runner, error) {
	if opts.MaxParallel <= 0 {
		opts.MaxParallel = 4
	}
	if opts.StaleDays <= 0 {
		opts.StaleDays = 30
	}
	if opts.SamplePaths <= 0 {
		opts.SamplePaths = 5
	}
	runID := time.Now().UTC().Format("20060102T150405Z")
	r := &Runner{
		opts:   opts,
		runID:  runID,
		budget: NewBudget(opts.Budget),
	}
	if opts.AuditDir != "" {
		log, err := audit.Open(opts.AuditDir, runID)
		if err != nil {
			return nil, fmt.Errorf("runner: open audit log: %w", err)
		}
		r.auditLog = log
	}
	if opts.UseTombstone && opts.TombstoneDir != "" {
		r.tombstone = tombstone.New(filepath.Join(opts.TombstoneDir))
	}
	return r, nil
}

// RunID is the timestamp identifier for this runner invocation.
func (r *Runner) RunID() string { return r.runID }

// AuditPath returns the audit-log file path ("" if disabled).
func (r *Runner) AuditPath() string {
	if r.auditLog == nil {
		return ""
	}
	return r.auditLog.Path()
}

// TombstoneRoot returns the per-run tombstone directory (or "" if disabled).
func (r *Runner) TombstoneRoot() string {
	if r.tombstone == nil {
		return ""
	}
	return filepath.Join(r.tombstone.BaseDir, r.runID)
}

// Close releases the audit log.
func (r *Runner) Close() error {
	if r.auditLog != nil {
		return r.auditLog.Close()
	}
	return nil
}

// Scan runs Detect → busy-check → Scan for each cleaner in parallel.
// Per-cleaner errors are captured on the ScanResult; the function
// returns nil unless ctx is cancelled.
func (r *Runner) Scan(ctx context.Context, cleaners []cleaner.Cleaner, progress func(ScanResult)) ([]ScanResult, error) {
	results := make([]ScanResult, len(cleaners))
	pool := NewPool(r.opts.MaxParallel)
	var mu sync.Mutex
	idxByName := map[string]int{}
	for i, c := range cleaners {
		idxByName[c.Name()] = i
		results[i].Cleaner = c
	}
	for _, c := range cleaners {
		c := c
		pool.Go(ctx, func(ctx context.Context) error {
			start := time.Now()
			res := ScanResult{Cleaner: c}
			defer func() {
				res.DurationMS = time.Since(start).Milliseconds()
				mu.Lock()
				results[idxByName[c.Name()]] = res
				mu.Unlock()
				if progress != nil {
					progress(res)
				}
			}()
			if !c.Detect(ctx) {
				res.NotInstalled = true
				return nil
			}
			if busy := fsutil.AnyProcessRunning(ctx, c.BusyProcesses()); busy != "" {
				res.BusyProcess = busy
				return nil
			}
			findings, err := c.Scan(ctx, cleaner.ScanOpts{
				StaleDays:   r.opts.StaleDays,
				SearchRoots: r.opts.SearchRoots,
				SamplePaths: r.opts.SamplePaths,
			})
			if err != nil {
				res.Err = err
				return nil
			}
			// Default Approved=true for every non-Risky finding so
			// non-interactive flows work without an explicit approval step.
			// Risky findings stay unapproved unless AutoApproveRisky is on
			// (set by --include-risky or --profile aggressive). The TUI
			// can override either way per finding.
			for i := range findings {
				findings[i].Cleaner = c.Name()
				if r.opts.AutoApproveRisky {
					findings[i].Approved = true
				} else {
					findings[i].Approved = !findings[i].Risky
				}
				res.TotalBytes += findings[i].Bytes
			}
			res.Findings = findings
			return nil
		})
	}
	_ = pool.Wait()
	if err := ctx.Err(); err != nil {
		return results, err
	}
	return results, nil
}

// Execute runs each cleaner's approved findings in parallel.
// The runner enforces a single-flight per cleaner Name() — useful for
// shell-out cleaners like `brew cleanup` that can't run concurrently
// with themselves.
func (r *Runner) Execute(ctx context.Context, scans []ScanResult, progress func(ExecResult)) []ExecResult {
	results := make([]ExecResult, len(scans))
	for i, s := range scans {
		results[i].ScanResult = s
	}
	pool := NewPool(r.opts.MaxParallel)
	var mu sync.Mutex

	// Filter scans down to those with at least one approved finding.
	type job struct {
		idx      int
		approved []cleaner.Finding
	}
	var jobs []job
	for i, s := range scans {
		if s.Err != nil || s.NotInstalled || s.BusyProcess != "" {
			continue
		}
		var ap []cleaner.Finding
		for _, f := range s.Findings {
			if f.Approved {
				ap = append(ap, f)
			}
		}
		if len(ap) > 0 {
			jobs = append(jobs, job{idx: i, approved: ap})
		}
	}

	var auditSink cleaner.AuditSink = audit.NopSink{}
	if r.auditLog != nil {
		auditSink = r.auditLog
	}

	for _, j := range jobs {
		j := j
		pool.Go(ctx, func(ctx context.Context) error {
			if r.budget.Exhausted() {
				mu.Lock()
				results[j.idx].Result = cleaner.Result{
					Cleaner: results[j.idx].Cleaner.Name(),
					Skipped: len(j.approved),
					Errors:  []error{errors.New("budget exhausted")},
					DryRun:  r.opts.DryRun,
				}
				mu.Unlock()
				return nil
			}
			tombstoneDir := ""
			useTomb := r.opts.UseTombstone && r.tombstone != nil
			if useTomb {
				tombstoneDir = r.TombstoneRoot()
			}
			start := time.Now()
			res, err := results[j.idx].Cleaner.Execute(ctx, j.approved, cleaner.ExecOpts{
				DryRun:       r.opts.DryRun,
				UseTombstone: useTomb,
				Tombstone:    tombstoneDir,
				AuditLog:     auditSink,
				RunID:        r.runID,
			})
			res.DurationMS = time.Since(start).Milliseconds()
			res.DryRun = r.opts.DryRun
			res.Tombstoned = useTomb
			if err != nil {
				res.Errors = append(res.Errors, err)
			}
			r.budget.Add(res.BytesFreed)
			mu.Lock()
			results[j.idx].Result = res
			mu.Unlock()
			if progress != nil {
				progress(results[j.idx])
			}
			return nil
		})
	}
	_ = pool.Wait()
	return results
}

// IsDryRun reports whether the runner is in dry-run mode.
func (r *Runner) IsDryRun() bool { return r.opts.DryRun }

// UsesTombstone reports whether the runner moves files to tombstone
// rather than deleting them outright.
func (r *Runner) UsesTombstone() bool { return r.opts.UseTombstone }

// SetDryRun flips the runner's dry-run state. Used by the TUI to let
// users toggle preview-vs-real mid-session without quitting.
func (r *Runner) SetDryRun(b bool) { r.opts.DryRun = b }

// SetSafe enables/disables the tombstone safety net mid-session.
// Lazily initialises the tombstone store the first time it's enabled.
func (r *Runner) SetSafe(b bool) {
	r.opts.UseTombstone = b
	if b && r.tombstone == nil && r.opts.TombstoneDir != "" {
		r.tombstone = tombstone.New(r.opts.TombstoneDir)
	}
}

// SweepTombstone removes runs older than retention. Safe to call
// before each new run.
func (r *Runner) SweepTombstone(retention time.Duration) ([]string, error) {
	if r.tombstone == nil {
		return nil, nil
	}
	return r.tombstone.Sweep(retention)
}
