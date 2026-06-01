// Package cleaner defines the Cleaner interface and shared types
// every cleanup module implements.
package cleaner

import (
	"context"
	"time"
)

type Category string

const (
	CategoryLangPkg   Category = "lang"
	CategoryIaC       Category = "iac"
	CategoryContainer Category = "container"
	CategorySystem    Category = "system"
	CategoryProject   Category = "project"
)

func (c Category) Title() string {
	switch c {
	case CategoryLangPkg:
		return "Language package caches"
	case CategoryIaC:
		return "IaC tool state"
	case CategoryContainer:
		return "Containers & VMs"
	case CategorySystem:
		return "System & apps"
	case CategoryProject:
		return "Project build artifacts"
	}
	return string(c)
}

// SamplePath is one concrete file or directory the cleaner would remove,
// surfaced to the user for transparency in the TUI review screen.
type SamplePath struct {
	Path         string
	Bytes        int64
	LastModified time.Time
}

// Finding is one unit of reclaimable space the cleaner has identified.
// A single cleaner can return many findings (e.g. terraform returns one
// per stale .terraform/ directory found).
type Finding struct {
	Cleaner      string
	Path         string
	Bytes        int64
	LastModified time.Time
	Reason       string
	Risky        bool
	// ShellOut signals this finding will be cleaned via a shell command
	// (e.g. `brew cleanup`, `docker system prune`) rather than file deletion.
	// Shell-out findings cannot be tombstoned.
	ShellOut    bool
	SamplePaths []SamplePath
	// Approved is set by the review/select phase. Runner only executes
	// findings whose Approved=true.
	Approved bool
}

// Result is the outcome of executing a cleaner's approved findings.
type Result struct {
	Cleaner    string
	BytesFreed int64 // best-effort; matches Finding.Bytes on success
	Findings   int
	Succeeded  int
	Failed     int
	Skipped    int
	DurationMS int64
	Errors     []error
	Tombstoned bool // true if findings were moved to tombstone rather than deleted
	DryRun     bool
}

// ScanOpts are the runtime options passed to Cleaner.Scan.
type ScanOpts struct {
	StaleDays   int
	SearchRoots []string // for IaC cleaners that walk a directory tree
	SamplePaths int      // how many top-N paths to attach per finding (0 = none)
}

// ExecOpts are the runtime options passed to Cleaner.Execute.
type ExecOpts struct {
	DryRun       bool
	UseTombstone bool
	// Tombstone is the absolute directory under which removed files
	// should be moved when UseTombstone=true. Cleaners that operate via
	// shell-out can ignore this; they have no choice but to unlink.
	Tombstone string
	// AuditLog receives one entry per file action. Never nil; runner
	// supplies a no-op when audit is disabled.
	AuditLog AuditSink
	RunID    string
}

// AuditSink decouples cleaners from the audit package to avoid an import
// cycle. The audit package implements this interface.
type AuditSink interface {
	Record(entry AuditEntry)
}

type AuditEntry struct {
	Timestamp  time.Time
	RunID      string
	Cleaner    string
	Path       string
	Bytes      int64
	DryRun     bool
	Tombstoned bool
	Success    bool
	Error      string
}

// Cleaner is implemented by every cleanup module.
//
// Lifecycle: Detect → (BusyProcesses check by runner) → Scan → Execute.
// Implementations should be safe to call concurrently with other
// cleaners; the runner enforces a single-flight per Name() during execute.
type Cleaner interface {
	Name() string
	Category() Category
	Description() string
	// Risky returns true if this cleaner deletes things that are
	// expensive to regenerate or could surprise a user. Risky cleaners
	// start unchecked in the TUI.
	Risky() bool
	// RiskReason returns a short human-readable note explaining *why*
	// this cleaner is risky (e.g. "may corrupt VM if running",
	// "forces multi-minute re-index"). Empty when Risky() is false.
	// Surfaced inline in `doctor` and `explain`; one flat "risky" tag
	// is too coarse — different cleaners are risky for different reasons.
	RiskReason() string
	// Detect returns true if the underlying tool is installed or the
	// target paths exist. Cleaners returning false are hidden in the TUI.
	Detect(ctx context.Context) bool
	// BusyProcesses returns process names whose presence (via pgrep)
	// should cause the runner to skip this cleaner. Empty slice = always
	// safe to run.
	BusyProcesses() []string
	// Scan is read-only; it computes findings and reclaimable bytes.
	Scan(ctx context.Context, opts ScanOpts) ([]Finding, error)
	// Execute deletes (or tombstones, or shells-out) the approved
	// findings. Honors opts.DryRun.
	Execute(ctx context.Context, findings []Finding, opts ExecOpts) (Result, error)
}
