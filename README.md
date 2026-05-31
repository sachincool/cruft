<h1 align="center">C&nbsp;R&nbsp;U&nbsp;F&nbsp;T</h1>
<p align="center"><em>decruft your dev laptop.</em></p>

<div align="center">
<pre>
~/Library/Caches   ~/.terraform   ~/.gradle/caches
~/Library/Developer/Xcode   ~/.cargo   ~/.npm   + a dozen more
                       │
                   c  r  u  f  t
                       │
     find it all  ·  tick what goes  ·  13.2 GB back
</pre>
</div>

<p align="center">one Go binary · 25 cleaners · deletes on confirm · <code>--safe</code> for 7-day undo · macOS</p>

<p align="center">
  <a href="https://github.com/sachincool/cruft/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/sachincool/cruft/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://github.com/sachincool/cruft/releases"><img alt="Release" src="https://img.shields.io/github/v/release/sachincool/cruft?display_name=tag"></a>
  <a href="LICENSE"><img alt="License: MIT" src="https://img.shields.io/badge/license-MIT-blue.svg"></a>
</p>

<p align="center">
  <img src="assets/demo.gif" alt="cruft's TUI: scan caches, review what's reclaimable and why, confirm a dry-run delete" width="900">
</p>

---

Your Mac says the disk is full. Not because of your code — because of the
caches every dev tool quietly piles up and never cleans. After a year of
real work you're probably sitting on:

- 5 GB of Xcode `DerivedData` from builds you'll never run again
- 3 GB of old Homebrew versions you already upgraded past
- a `.terraform/` and `.terragrunt-cache/` in each of a dozen infra repos you cloned once
- 1 GB of npm cache from a project you opened twice
- a JetBrains index for a project you deleted months ago

You know most of it is safe to delete. The problem is remembering where it
all lives, and knowing which bits cost you a 15-minute rebuild if you guess wrong.

`cruft` knows. It scans for all of it at once, shows you what it found and why
each thing is safe to remove, and deletes only what you tick.

> On my M1, the default profile cleared **13.2 GB across 19 cleaners in 11 seconds.** Adding `--include-risky` turned up another 620 MB.

## Install

**Homebrew** — recommended:

```sh
brew install sachincool/tap/cruft
```

**Go** — build from source:

```sh
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

### Shell completions

cruft completes both its subcommands and cleaner names — `cruft run <Tab>`
suggests `npm`, `docker`, `xcode-derived`, and the rest.

```sh
source <(cruft completion zsh)      # zsh  — add to ~/.zshrc
source <(cruft completion bash)     # bash — add to ~/.bashrc
cruft completion fish | source      # fish
```

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

## Is this safe?

Short answer: yes — and you can always check its work.

- **It only touches paths it already knows.** Every cleaner lists the exact
  directories it owns. cruft follows symlinks to where they really point,
  then refuses to delete anything that isn't on the list. No globs, no
  "delete everything under here."
- **The risky stuff is opt-in.** Six cleaners are marked *risky* because what
  they delete takes a while to come back (a JetBrains re-index, an Xcode
  archive). They never run by default, and each one tells you why it's risky.
- **It waits for your tools.** If `node` is running, the npm cleaner skips
  this round. If Xcode is open, DerivedData is left alone — so nothing gets
  pulled out from under a running build.
- **Every run leaves a receipt.** Each action, dry-runs included, writes a
  line of JSON to `~/.local/share/cruft/runs/`. Read it back any time with
  `cruft last` or `cruft history`.
- **You can undo it.** Run with `--safe` and deletions go to a recycle bin for
  7 days — `cruft restore <run-id>` brings them back. It's off by default
  because most of what cruft cleans just regenerates the next time you use the tool.

## Profiles

A profile is just how much cruft is willing to clean:

```sh
cruft --profile conservative   # only caches that come back in seconds
cruft --profile balanced       # default — everything that's safe
cruft --profile aggressive     # also cleans the risky stuff, automatically
```

Even on the default profile, risky cleaners still show up in the results so
you can see what's there. They just don't get deleted unless you opt in.

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
