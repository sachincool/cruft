<h1 align="center">C&nbsp;R&nbsp;U&nbsp;F&nbsp;T&nbsp;&nbsp;<sub>░▒▓</sub></h1>
<p align="center"><em>decruft your dev laptop.</em></p>
<p align="center">one Go binary · 25 cleaners · deletes on confirm · <code>--safe</code> for 7-day undo · macOS</p>

<p align="center">
  <a href="https://github.com/sachincool/cruft/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/sachincool/cruft/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://github.com/sachincool/cruft/releases"><img alt="Release" src="https://img.shields.io/github/v/release/sachincool/cruft?display_name=tag"></a>
  <a href="LICENSE"><img alt="License: MIT" src="https://img.shields.io/badge/license-MIT-blue.svg"></a>
</p>

<p align="center">
  <img src="assets/demo.gif" alt="cruft TUI scanning 19 cleaners and reporting 13.2 GB of reclaimable space" width="900">
</p>

---

You probably have:

- 5 GB of Xcode `DerivedData` from last year
- 3 GB of old Homebrew formula versions
- a `.terraform/` for a repo you don't recognise
- 1 GB of npm cache from a project you ran twice
- a JetBrains index for a project you deleted

`cruft` finds it all in parallel, shows you what it found and why, and deletes what you tick.

> On my M1, balanced profile: **13.2 GB across 19 cleaners in 11s.** With `--include-risky`: +620 MB more.

## Install

```sh
brew install sachincool/tap/cruft     # macOS, recommended
go install github.com/sachincool/cruft/cmd/cruft@latest
```

5.6 MB binary, single file, zero runtime deps.

## Use it

```sh
cruft                          # interactive TUI: pick, confirm, done
cruft run --all                # non-interactive, deletes on confirm
cruft run --all --dry-run      # preview only, changes nothing
cruft run --all --safe         # delete via 7-day recycle (`cruft restore` works)

cruft doctor                   # what's installed, what's busy
cruft explain xcode-derived    # what one cleaner actually does
cruft glossary                 # define every term cruft uses
cruft last                     # what the last run did
```

Three keys you'll use in the TUI:

| key | what |
|---|---|
| `space` | toggle the cleaner under the cursor |
| `?` | what does this cleaner do? |
| `d` / `s` | flip `--dry-run` / `--safe` mid-session |

## What it cleans

| | |
|---|---|
| **Language caches** | npm · pnpm · yarn · pip · cargo · Go module cache · gem · gradle · maven |
| **IaC state** | terraform · terragrunt · pulumi |
| **Containers** | docker · docker volumes · colima |
| **macOS apps** | Homebrew · Xcode DerivedData · Xcode archives · unavailable simulators · Library/Caches (allowlist) · VS Code · JetBrains caches · JetBrains system · Slack · Trash |

Risky cleaners (Docker volumes, JetBrains system indexes, Xcode archives, Trash, …) stay opt-in. Each one prints its own one-line reason in `cruft doctor` and `cruft explain`. Surface them with `--include-risky` or `--profile aggressive`.

## What it doesn't do

- Touch your code, your dotfiles, or anything outside a cleaner's declared paths.
- Run as root.
- Follow a symlink out of an allowed directory.
- Delete a cleaner's cache while the matching tool is running (`npm` skips if `node` is running, `terraform` skips if `terraform` is, etc.).
- Cover Linux or Windows yet.

## Safety, plainly

- **Whitelist-only paths.** Every cleaner declares the directories it owns. `fsutil.SafeRemove` resolves symlinks then verifies the target is under one of them. Anything else is refused with a logged error.
- **Risky cleaners are opt-in.** Six cleaners are marked risky for a reason that's printed inline (see `cruft glossary risky`). None of them run on the default `balanced` profile.
- **Process guards.** Each cleaner names processes that block it. If `gopls` is running, the Go module cleaner sits this run out.
- **Audit log.** Every action — including dry-runs — writes one JSON line to `~/.local/share/cruft/runs/<run-id>.jsonl`. Reads back via `cruft history` / `cruft last`.
- **`--safe` for undo.** Off by default because most of what cruft cleans regenerates on next use (npm/yarn/pip cache, Xcode DerivedData, brew cleanup). Turn it on and deletions move to `~/.local/share/cruft/tombstone/<run-id>`; `cruft restore <run-id>` brings them back for the next 7 days.

## Profiles

```sh
cruft --profile conservative   # only caches that regenerate in seconds
cruft --profile balanced       # default — everything safe
cruft --profile aggressive     # also auto-approves risky cleaners
```

On `balanced` (the default), risky cleaners are still **scanned and surfaced** in the summary so you can see what's available — they just don't run unless you opt in.

## JSON output

```sh
cruft run --all --dry-run --json | jq '[.results[] | select(.bytes_freed > 0)] | sort_by(-.bytes_freed)'
```

Every subcommand that prints a summary takes `--json`. Schema lives in [`docs/json-output.md`](docs/json-output.md). Useful for before/after disk reports across a fleet.

## Why not just `rm -rf`?

Because there are 25 different paths spread across `~/Library/`, `~/.cache/`, `~/.cargo/`, `~/.gradle/`, `~/Library/Developer/Xcode/`, `~/.colima/`, and a dozen more, and you don't want to type them wrong at midnight. cruft has them registered, declares which are risky, knows which tools have to be quiet first, and writes a receipt you can `jq`.

## Develop

```sh
make check        # gofmt, vet, tests
make build        # bin/cruft
make snapshot     # local goreleaser snapshot
```

Add a cleaner: copy any file under `internal/cleaner/<category>/`, implement the four-method interface (`Detect` / `BusyProcesses` / `Scan` / `Execute`), call `cleaner.Register(&yours{})` in `init()`. Open a PR with a one-line note on what the cleaner owns and whether it's risky (and why).

## Status

macOS only. ~4,700 lines of Go across 58 files. 25 cleaners. Used on the maintainer's laptop. PRs welcome — especially Linux support, more cleaners, and tests with real fixtures.

License: MIT
