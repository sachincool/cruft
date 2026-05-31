package system

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/sachincool/cruft/internal/cleaner"
	"github.com/sachincool/cruft/internal/fsutil"
)

type xcodeSimsCleaner struct{}

func init() { cleaner.Register(&xcodeSimsCleaner{}) }

func (xcodeSimsCleaner) Name() string               { return "xcode-simulators" }
func (xcodeSimsCleaner) Category() cleaner.Category { return cleaner.CategorySystem }
func (xcodeSimsCleaner) Description() string {
	return "Removes Xcode simulators marked 'unavailable' (left over from old SDKs). Never deletes active simulators — your login state and app data are safe."
}
func (xcodeSimsCleaner) Risky() bool             { return false }
func (xcodeSimsCleaner) RiskReason() string      { return "" }
func (xcodeSimsCleaner) BusyProcesses() []string { return []string{"Simulator", "Xcode"} }

func (xcodeSimsCleaner) Detect(ctx context.Context) bool { return fsutil.CommandExists("xcrun") }

func (xcodeSimsCleaner) Scan(ctx context.Context, opts cleaner.ScanOpts) ([]cleaner.Finding, error) {
	out, err := exec.CommandContext(ctx, "xcrun", "simctl", "list", "devices", "unavailable").Output()
	if err != nil {
		return nil, nil
	}
	// Count entries — we can't easily get bytes without filesystem
	// inspection, so estimate at ~1 GB per unavailable simulator
	// (typical disk footprint). Conservative; user sees count.
	var count int
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") || strings.HasSuffix(line, ":") {
			continue
		}
		if strings.Contains(line, "(unavailable") {
			count++
		}
	}
	if count == 0 {
		return nil, nil
	}
	return []cleaner.Finding{{
		Path:     "xcode:unavailable-simulators",
		Bytes:    int64(count) * 1_000_000_000, // ~1 GB estimate per sim
		Reason:   "unavailable simulators (rough estimate; ~1 GB/sim)",
		ShellOut: true,
	}}, nil
}

func (xcodeSimsCleaner) Execute(ctx context.Context, findings []cleaner.Finding, opts cleaner.ExecOpts) (cleaner.Result, error) {
	res := cleaner.Result{Cleaner: "xcode-simulators", Findings: len(findings), DryRun: opts.DryRun}
	for _, f := range findings {
		entry := cleaner.AuditEntry{
			Timestamp: time.Now(), RunID: opts.RunID, Cleaner: "xcode-simulators",
			Path: f.Path, Bytes: f.Bytes, DryRun: opts.DryRun,
		}
		if opts.DryRun {
			entry.Success = true
			opts.AuditLog.Record(entry)
			res.Succeeded++
			res.BytesFreed += f.Bytes
			continue
		}
		if err := exec.CommandContext(ctx, "xcrun", "simctl", "delete", "unavailable").Run(); err != nil {
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
