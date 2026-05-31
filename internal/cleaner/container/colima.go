package container

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/sachincool/cruft/internal/cleaner"
	"github.com/sachincool/cruft/internal/fsutil"
)

type colimaCleaner struct{}

func init() { cleaner.Register(&colimaCleaner{}) }

func (colimaCleaner) Name() string               { return "colima" }
func (colimaCleaner) Category() cleaner.Category { return cleaner.CategoryContainer }
func (colimaCleaner) Description() string {
	return "Trim the Colima VM disk image. Requires Colima to be STOPPED — running this on a live VM can corrupt the image. We refuse if `colima status` reports running."
}
func (colimaCleaner) Risky() bool { return true }
func (colimaCleaner) RiskReason() string {
	return "trimming a running VM image can corrupt it; refuses if Colima is up"
}
func (colimaCleaner) BusyProcesses() []string { return []string{"colima"} }

func (colimaCleaner) Detect(ctx context.Context) bool {
	return fsutil.CommandExists("colima")
}

func (colimaCleaner) Scan(ctx context.Context, opts cleaner.ScanOpts) ([]cleaner.Finding, error) {
	// Disk image size (if we can find it). Standard location:
	// ~/.colima/_lima/colima/diffdisk
	candidates := []string{
		"~/.colima/_lima/colima/diffdisk",
		"~/.colima/default/diffdisk",
	}
	for _, c := range candidates {
		path := fsutil.Expand(c)
		if !fsutil.Exists(path) {
			continue
		}
		s, err := fsutil.Size(ctx, path, 0)
		if err != nil || s.Bytes == 0 {
			continue
		}
		// We can't actually estimate reclaimable bytes without trimming
		// — report the total disk image size and let the user decide.
		return []cleaner.Finding{{
			Path:     path,
			Bytes:    s.Bytes,
			Reason:   "colima VM disk image (trim while stopped)",
			Risky:    true,
			ShellOut: true,
		}}, nil
	}
	return nil, nil
}

func (colimaCleaner) Execute(ctx context.Context, findings []cleaner.Finding, opts cleaner.ExecOpts) (cleaner.Result, error) {
	res := cleaner.Result{Cleaner: "colima", Findings: len(findings), DryRun: opts.DryRun}
	for _, f := range findings {
		entry := cleaner.AuditEntry{
			Timestamp: time.Now(), RunID: opts.RunID, Cleaner: "colima",
			Path: f.Path, Bytes: f.Bytes, DryRun: opts.DryRun,
		}
		if opts.DryRun {
			entry.Success = true
			opts.AuditLog.Record(entry)
			res.Succeeded++
			continue
		}
		// Hard refuse if colima is up — trimming a live image is destructive.
		status, _ := exec.CommandContext(ctx, "colima", "status").CombinedOutput()
		if strings.Contains(string(status), "running") || strings.Contains(string(status), "Running") {
			err := fmt.Errorf("colima is running; stop it first with `colima stop`")
			entry.Error = err.Error()
			opts.AuditLog.Record(entry)
			res.Failed++
			res.Errors = append(res.Errors, err)
			continue
		}
		// Use limactl to sparsify the disk image (Colima ships with limactl).
		if err := exec.CommandContext(ctx, "limactl", "disk", "sparsify", "colima").Run(); err != nil {
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
