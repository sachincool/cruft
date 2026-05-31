package system

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/sachincool/cruft/internal/cleaner"
	"github.com/sachincool/cruft/internal/fsutil"
)

type brewCleaner struct{}

func init() { cleaner.Register(&brewCleaner{}) }

func (brewCleaner) Name() string               { return "homebrew" }
func (brewCleaner) Category() cleaner.Category { return cleaner.CategorySystem }
func (brewCleaner) Description() string {
	return "Homebrew: removes old formula versions and the download cache. Loses `brew switch` rollback for those versions."
}
func (brewCleaner) Risky() bool             { return false }
func (brewCleaner) RiskReason() string      { return "" }
func (brewCleaner) BusyProcesses() []string { return []string{"brew"} }

func (brewCleaner) Detect(ctx context.Context) bool { return fsutil.CommandExists("brew") }

func (brewCleaner) Scan(ctx context.Context, opts cleaner.ScanOpts) ([]cleaner.Finding, error) {
	// `brew cleanup --dry-run -s` lists what would be removed. The last
	// line is like: "==> This operation would free approximately 1.2GB of disk space."
	out, err := exec.CommandContext(ctx, "brew", "cleanup", "--dry-run", "-s").CombinedOutput()
	if err != nil {
		return nil, nil
	}
	bytes := parseBrewBytes(string(out))
	if bytes <= 0 {
		return nil, nil
	}
	return []cleaner.Finding{{
		Path:     "homebrew:cleanup",
		Bytes:    bytes,
		Reason:   "old formula versions + download cache",
		ShellOut: true,
	}}, nil
}

func (brewCleaner) Execute(ctx context.Context, findings []cleaner.Finding, opts cleaner.ExecOpts) (cleaner.Result, error) {
	res := cleaner.Result{Cleaner: "homebrew", Findings: len(findings), DryRun: opts.DryRun}
	for _, f := range findings {
		entry := cleaner.AuditEntry{
			Timestamp: time.Now(), RunID: opts.RunID, Cleaner: "homebrew",
			Path: f.Path, Bytes: f.Bytes, DryRun: opts.DryRun,
		}
		if opts.DryRun {
			entry.Success = true
			opts.AuditLog.Record(entry)
			res.Succeeded++
			res.BytesFreed += f.Bytes
			continue
		}
		if err := exec.CommandContext(ctx, "brew", "cleanup", "-s").Run(); err != nil {
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

// parseBrewBytes scrapes brew's "would free approximately X" line.
func parseBrewBytes(output string) int64 {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "would free") {
			continue
		}
		// "==> This operation would free approximately 1.2GB of disk space."
		idx := strings.Index(line, "approximately")
		if idx < 0 {
			continue
		}
		rest := strings.TrimSpace(line[idx+len("approximately"):])
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}
		return parseBrewSize(fields[0])
	}
	return 0
}

func parseBrewSize(s string) int64 {
	multipliers := []struct {
		suffix string
		mul    float64
	}{{"TB", 1e12}, {"GB", 1e9}, {"MB", 1e6}, {"KB", 1e3}, {"B", 1}}
	for _, m := range multipliers {
		if strings.HasSuffix(s, m.suffix) {
			num := strings.TrimSuffix(s, m.suffix)
			f, err := strconv.ParseFloat(num, 64)
			if err != nil {
				return 0
			}
			return int64(f * m.mul)
		}
	}
	return 0
}
