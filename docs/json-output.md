# JSON output

Use `--json` with non-interactive commands when CRUFT is part of scripts, reports, or workstation health checks.

```sh
cruft run --all --json > cruft-report.json
cruft list --json
cruft history --json
```

## `cruft run --all --json`

Example shape:

```json
{
  "run_id": "20260528T120000Z",
  "audit_log": "/Users/alice/.local/share/cruft/runs/20260528T120000Z.jsonl",
  "tombstone": "/Users/alice/.local/share/cruft/tombstone/20260528T120000Z",
  "results": [
    {
      "cleaner": "npm",
      "findings": 1,
      "bytes_freed": 734003200,
      "succeeded": 1,
      "failed": 0,
      "duration_ms": 428
    }
  ]
}
```

Field notes:

- `bytes_freed` is best-effort and equals reclaimable bytes during dry-run.
- `not_installed` means the cleaner's underlying tool/path was absent.
- `busy_process` means CRUFT skipped the cleaner because a related process was running.
- `errors` contains per-cleaner execution errors.
- Dry-run reports still include `succeeded` for findings that would be removed.

## Audit JSONL

Each line in the audit log is one action:

```json
{
  "timestamp": "2026-05-28T12:00:01Z",
  "run_id": "20260528T120000Z",
  "cleaner": "vscode",
  "path": "/Users/alice/Library/Application Support/Code/Cache",
  "bytes": 104857600,
  "dry_run": true,
  "tombstoned": false,
  "success": true,
  "error": ""
}
```

Treat the schema as stable within a major version. New optional fields may be added over time.
