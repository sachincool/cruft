# Safety model

What stops cruft from deleting something you'd miss.

## Lifecycle

```text
Detect tool/path
  ↓
Check busy processes  (skip if a relevant process is running)
  ↓
Scan: compute findings, sizes, and sample paths   (read-only)
  ↓
Approval                                          (profile + TUI / flags)
  ↓
Execute approved findings                         (--dry-run prints, doesn't act)
  ↓
Write audit log
  ↓
Either delete outright, or (with --safe) move into the recycle bin
```

## Defaults

- `--dry-run` is **off** by default. `cruft run --all` deletes on confirm.
  Pass `--dry-run` to preview; in the TUI press `d`.
- `--safe` is **off** by default. Most of what cruft cleans regenerates
  on next use (npm cache, Xcode DerivedData, brew downloads), so an
  always-on recycle bin would just consume disk for no real benefit.
  Pass `--safe` to route deletions through a 7-day recycle; in the TUI
  press `s`.
- The `balanced` profile excludes risky cleaners from auto-approval but
  **still scans them** so their reclaim is visible in the summary,
  with the exact command to opt in.
- Every risky cleaner must declare a one-line `RiskReason()` shown in
  `cruft doctor`, `cruft explain`, the TUI help panel (`?`), and the
  summary's risky-offer block.
- Cleaners skip work when their `BusyProcesses()` are running (`npm`
  skips if `node` is running, `terraform` skips if `terraform` is, etc.).

## Filesystem deletion rules

Path-based cleaners declare their owned paths. At execution time:

1. Each finding's path is expanded (`~`, env vars resolved).
2. `filepath.EvalSymlinks` resolves any symlinks in the path.
3. The resolved path is checked against the cleaner's whitelist; if it
   isn't under one of the allowed prefixes, the deletion is refused
   with a logged error.
4. cruft refuses to run as root (`EUID==0`).
5. Without `--safe`: `os.RemoveAll` on the resolved path.
6. With `--safe`: the file is `os.Rename`d into the recycle bin
   (atomic on the same filesystem), an entry is appended to that
   run's `manifest.jsonl`, and the audit log records the move.

A symlink inside an allowed directory cannot be used to redirect the
deletion to a target outside that directory, because the resolution
step happens before the whitelist check.

## Recycle bin and restore

Only with `--safe`. Files move to:

```text
~/.local/share/cruft/tombstone/<run-id>/files/<original-absolute-path>
```

The per-run `manifest.jsonl` records source → destination for every
moved entry. `cruft restore <run-id>` walks it in reverse:

- If the original path is already occupied (e.g. the app rebuilt its
  cache), that entry is skipped and reported, never clobbered.
- Otherwise the file is renamed back to its original location.
- If everything restores cleanly, the empty per-run directory is removed.

After `--tombstone-days` (default 7), the next cruft invocation sweeps
expired recycle directories.

## Shell-out cleaners

A few cleaners (`homebrew`, `docker`, `docker-volumes`, `xcode-simulators`)
invoke the upstream tool's own cleanup (`brew cleanup`, `docker system
prune`, `xcrun simctl delete unavailable`). These can't be routed
through the recycle bin — the upstream tool unlinks files directly.
The audit log still records what happened, but `cruft restore` won't
bring them back. Each such cleaner's `Description()` calls this out.

## Audit log

One JSON line per finding, written to:

```text
~/.local/share/cruft/runs/<run-id>.jsonl
```

Fields: `timestamp`, `run_id`, `cleaner`, `path`, `bytes`, `dry_run`,
`tombstoned`, `success`, `error`. Read it back with `cruft history`
or `cruft last`, or grep it with `jq`.

## What cruft will never do

- Delete a path that isn't under a cleaner's declared whitelist.
- Follow a symlink outside that whitelist.
- Run as root.
- Run a risky cleaner without `--include-risky` (or `--profile aggressive`,
  or an explicit toggle in the TUI).
- Delete a cache while its owning tool is running.
