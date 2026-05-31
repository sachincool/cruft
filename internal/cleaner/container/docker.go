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

type dockerCleaner struct{}

func init() { cleaner.Register(&dockerCleaner{}) }

func (dockerCleaner) Name() string               { return "docker" }
func (dockerCleaner) Category() cleaner.Category { return cleaner.CategoryContainer }
func (dockerCleaner) Description() string {
	return "Docker dangling images, stopped containers, unused networks, and build cache. Volumes are NOT touched (see docker-volumes)."
}
func (dockerCleaner) Risky() bool             { return false }
func (dockerCleaner) RiskReason() string      { return "" }
func (dockerCleaner) BusyProcesses() []string { return nil }

func (dockerCleaner) Detect(ctx context.Context) bool {
	if !fsutil.CommandExists("docker") {
		return false
	}
	// docker info exits non-zero if the daemon isn't running.
	cmd := exec.CommandContext(ctx, "docker", "info")
	return cmd.Run() == nil
}

type dockerDFItem struct {
	Type        string
	Reclaimable string
}

func (dockerCleaner) Scan(ctx context.Context, opts cleaner.ScanOpts) ([]cleaner.Finding, error) {
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
		// Reclaimable looks like "1.2GB" or "850MB (50%)".
		bytes += parseDockerSize(item.Reclaimable)
	}
	if bytes <= 0 {
		return nil, nil
	}
	return []cleaner.Finding{{
		Path:     "docker:reclaimable",
		Bytes:    bytes,
		Reason:   "dangling images, stopped containers, unused networks, build cache",
		ShellOut: true,
	}}, nil
}

func (dockerCleaner) Execute(ctx context.Context, findings []cleaner.Finding, opts cleaner.ExecOpts) (cleaner.Result, error) {
	res := cleaner.Result{Cleaner: "docker", Findings: len(findings), DryRun: opts.DryRun}
	for _, f := range findings {
		entry := cleaner.AuditEntry{
			Timestamp: time.Now(), RunID: opts.RunID, Cleaner: "docker",
			Path: f.Path, Bytes: f.Bytes, DryRun: opts.DryRun,
		}
		if opts.DryRun {
			entry.Success = true
			opts.AuditLog.Record(entry)
			res.Succeeded++
			res.BytesFreed += f.Bytes
			continue
		}
		// `docker system prune -f`: removes stopped containers, dangling
		// images, unused networks, build cache. Does NOT touch volumes.
		if err := exec.CommandContext(ctx, "docker", "system", "prune", "-f").Run(); err != nil {
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

// parseDockerSize converts "1.2GB", "850MB (50%)" etc. to bytes.
func parseDockerSize(s string) int64 {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, " "); i > 0 {
		s = s[:i]
	}
	// Strip percentage in parens if no space, e.g. "1.2GB(50%)".
	if i := strings.Index(s, "("); i > 0 {
		s = s[:i]
	}
	if s == "" || s == "0B" {
		return 0
	}
	multipliers := []struct {
		suffix string
		mul    float64
	}{{"TB", 1e12}, {"GB", 1e9}, {"MB", 1e6}, {"kB", 1e3}, {"B", 1}}
	for _, m := range multipliers {
		if strings.HasSuffix(s, m.suffix) {
			num := strings.TrimSuffix(s, m.suffix)
			f := parseFloat(num)
			return int64(f * m.mul)
		}
	}
	return 0
}

func parseFloat(s string) float64 {
	var out float64
	var frac float64
	var fracPow float64 = 1
	inFrac := false
	for _, r := range s {
		if r == '.' {
			inFrac = true
			continue
		}
		if r < '0' || r > '9' {
			break
		}
		if inFrac {
			fracPow *= 10
			frac = frac*10 + float64(r-'0')
		} else {
			out = out*10 + float64(r-'0')
		}
	}
	return out + frac/fracPow
}
