package container

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"

	"github.com/sachincool/cruft/internal/cleaner"
	"github.com/sachincool/cruft/internal/fsutil"
)

type dockerVolumesCleaner struct{}

func init() { cleaner.Register(&dockerVolumesCleaner{}) }

func (dockerVolumesCleaner) Name() string               { return "docker-volumes" }
func (dockerVolumesCleaner) Category() cleaner.Category { return cleaner.CategoryContainer }
func (dockerVolumesCleaner) Description() string {
	return "Anonymous Docker volumes not referenced by any container. These often hold database files — only enable if you know none of your stopped containers will be restarted."
}
func (dockerVolumesCleaner) Risky() bool { return true }
func (dockerVolumesCleaner) RiskReason() string {
	return "anonymous volumes often hold database/state — gone for good when pruned"
}
func (dockerVolumesCleaner) BusyProcesses() []string { return nil }

func (dockerVolumesCleaner) Detect(ctx context.Context) bool {
	if !fsutil.CommandExists("docker") {
		return false
	}
	return exec.CommandContext(ctx, "docker", "info").Run() == nil
}

func (dockerVolumesCleaner) Scan(ctx context.Context, opts cleaner.ScanOpts) ([]cleaner.Finding, error) {
	out, err := exec.CommandContext(ctx, "docker", "system", "df", "--format", "{{json .}}").Output()
	if err != nil {
		return nil, nil
	}
	var bytes int64
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var item dockerDFItem
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}
		if item.Type == "Local Volumes" {
			bytes = parseDockerSize(item.Reclaimable)
			break
		}
	}
	if bytes <= 0 {
		return nil, nil
	}
	return []cleaner.Finding{{
		Path:     "docker:unused-volumes",
		Bytes:    bytes,
		Reason:   "anonymous volumes not in use by any container",
		Risky:    true,
		ShellOut: true,
	}}, nil
}

func (dockerVolumesCleaner) Execute(ctx context.Context, findings []cleaner.Finding, opts cleaner.ExecOpts) (cleaner.Result, error) {
	res := cleaner.Result{Cleaner: "docker-volumes", Findings: len(findings), DryRun: opts.DryRun}
	for _, f := range findings {
		entry := cleaner.AuditEntry{
			Timestamp: time.Now(), RunID: opts.RunID, Cleaner: "docker-volumes",
			Path: f.Path, Bytes: f.Bytes, DryRun: opts.DryRun,
		}
		if opts.DryRun {
			entry.Success = true
			opts.AuditLog.Record(entry)
			res.Succeeded++
			res.BytesFreed += f.Bytes
			continue
		}
		if err := exec.CommandContext(ctx, "docker", "volume", "prune", "-f").Run(); err != nil {
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
