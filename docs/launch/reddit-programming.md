# /r/programming

## Title
cruft: a macOS CLI for the 13 GB of dev caches you forgot existed

## Body
I built a small Go tool to solve a problem I kept hitting: my dev laptop runs out of disk every few months, and the offending bytes are scattered across paths I can never remember. `~/.cargo/registry/cache`, `~/Library/Developer/Xcode/DerivedData`, `~/Library/Caches/JetBrains`, every `.terraform/` dir under every infra repo I've ever cloned.

`cruft` ships 25 cleaners that know these paths so I don't have to.

Repo: https://github.com/sachincool/cruft

### What's covered
- Language package caches: npm, pnpm, yarn, pip, cargo, Go modcache, gem, gradle, maven
- IaC state: terraform (walks from a root and finds every `.terraform/`), terragrunt cache, pulumi
- Containers: Docker, Colima
- macOS apps: Homebrew downloads, Xcode DerivedData + archives + simulators, VS Code caches, JetBrains caches and system dirs, Slack, Trash, `~/Library/Caches`

Real number from my M1: 13.2 GB freed across 19 cleaners in 11 seconds.

### Safety model
This is the part I'd want to know about before running someone else's deletion tool, so:

1. **Whitelist-only.** Each cleaner declares its target paths explicitly. There are no glob patterns and no "rm everything under X". If a path doesn't match the whitelist it doesn't get touched.
2. **Symlink-escape protection.** Before deleting, each path is resolved (`filepath.EvalSymlinks`) and the resolved path is checked against the whitelist root. If a symlink in the tree points outside, that entry is skipped.
3. **Process awareness.** Each cleaner can declare conflicting processes. If `Docker.app` is running, the Docker cleaner is skipped with a printed reason. Same for Xcode, the JetBrains IDEs, etc.
4. **Risky cleaners are opt-in.** A subset of cleaners (the ones that can corrupt a VM if it's running, or force a multi-minute re-index next launch) are off by default. `--include-risky` turns them on, and each carries a per-cleaner reason string that prints at runtime.
5. **Recycle bin is opt-in, not default.** `--safe` routes deletions through a 7-day recycle bin under the cache dir. The default is straight unlink — I want a small fast tool, not a backup system. If you want the safety net, ask for it.
6. **Audit log.** Every run writes a JSONL log: cleaner name, path, bytes freed, errors. Good for "wait, what did I just delete".

### How it's built
~4,700 LOC of Go, single 5.6 MB binary, no CGo. Bubbletea TUI for interactive use; `cruft run --all` for scripts. Cleaners implement a small interface (Name, Paths, Scan, Clean, IsRisky, ConflictingProcesses) so adding one is mechanical.

### Honest disclaimers
- macOS only right now. The Xcode / Library / Homebrew bits would need to be factored out for Linux.
- Brand new, written by one person (me). Use `--safe` for the first run if that makes you nervous.
- Not on the roadmap: GUI, daemon, telemetry, scheduling.

### Install
```
go install github.com/sachincool/cruft/cmd/cruft@latest
```
Homebrew tap coming once the formula passes review: `brew install sachincool/tap/cruft`.

Happy to take feedback on the cleaner taxonomy or the safety model — both are easier to fix now than later.
