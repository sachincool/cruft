# Contributing to cruft

cruft deletes developer-machine state. Every contribution should keep the
tool predictable, safe, and easy to inspect.

## Local checks

```sh
make check        # gofmt + tests + go vet
make build        # bin/cruft
make snapshot     # local goreleaser snapshot (validates release config)
```

CI runs the same checks on darwin and linux for every PR.

## The Cleaner interface (9 methods)

Every cleaner implements `cleaner.Cleaner` in `internal/cleaner/cleaner.go`:

| method | does what |
|---|---|
| `Name() string` | unique slug, used in `cruft list`, `--name` args, etc. |
| `Category() Category` | `lang` / `iac` / `container` / `system` |
| `Description() string` | one-paragraph plain-language doc shown in `cruft explain` and the TUI's `?` panel |
| `Risky() bool` | true if findings stay unapproved by default |
| `RiskReason() string` | one-line explanation of *why* it's risky; required when `Risky()` returns true |
| `Detect(ctx) bool` | true if the underlying tool is installed and the cleaner is applicable |
| `BusyProcesses() []string` | process names (matched by `pgrep -x`) that, when running, skip the cleaner this run |
| `Scan(ctx, opts) ([]Finding, error)` | read-only; compute findings and reclaimable bytes |
| `Execute(ctx, findings, opts) (Result, error)` | act on the approved subset; honour `opts.DryRun` |

For most cleaners you can skip writing the full interface and embed
`cleaner.PathCleaner` from `internal/cleaner/common.go` — see any file
under `internal/cleaner/lang/` for the four-line registration form.

## Cleaner PR checklist

Include these in the PR description:

- exact paths or shell commands the cleaner owns
- regeneration cost (seconds, minutes, hours)
- busy processes that should make it skip
- whether it's risky, and if so the one-line `RiskReason()` string
- `cruft explain <name>` output (paste it)
- `cruft run <name> --dry-run` output on your machine, with sensitive paths redacted

A worked example is in [`docs/cleaner-anatomy.md`](docs/cleaner-anatomy.md).

## Design rules

These are the invariants. Break one only with strong justification.

- **Whitelist-only deletion.** Every finding's path must resolve (through
  symlinks) to somewhere under the cleaner's declared paths.
  `internal/fsutil/safe_remove.go` enforces this; don't bypass it.
- **Refuse to run as root.** Already enforced; don't add a flag to disable.
- **Process awareness.** If your cleaner touches a cache the running tool
  might be writing to, list that tool in `BusyProcesses()`.
- **Risky must be specific.** A flat "risky" tag is too coarse — your
  `RiskReason()` must be one concrete sentence ("forces multi-minute
  re-index", "may corrupt VM if running"), not a vague hedge.
- **Shell-out cleaners can't be recycled.** If your cleaner shells out to
  `brew cleanup` / `docker prune` / similar, the upstream tool unlinks
  files directly — `--safe` recycle has nothing to bury. Note this in
  your `Description()`.
- **Audit every action.** Every deletion (including dry-run) emits one
  JSON line via `opts.AuditLog.Record(...)`. The runner wires this up;
  use it.

## What cruft is *not* trying to be

- Not a generic macOS cleaner. We don't touch system caches like
  `~/Library/Saved Application State`, font caches, or anything Apple
  signs. Use BleachBit or similar.
- Not cross-platform yet. Linux/WSL PRs are welcome but should land
  behind a build tag rather than runtime `if runtime.GOOS == ...` branching.
- Not destructive by default. We delete on confirm; we do not auto-prune.

## Voice & code style

- Code: standard `gofmt`. Names should match what the user types
  (`xcode-derived`, not `xcode_data` or `xcd`).
- Doc strings: short, plain language, no marketing copy.
- `Description()` and `RiskReason()` end up in user-facing output —
  write them like you're explaining to a stranger debugging at 1 a.m.

## Reporting bugs

Use the bug report template. Reproductions beat descriptions; redact
private paths before pasting.

## Code of conduct

By participating you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).
