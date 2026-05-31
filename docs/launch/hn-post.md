# Show HN

## Title
Show HN: Cruft – a CLI that finds the 13 GB of dev caches on your Mac

## Body
I write code on a laptop with a 256 GB disk. Every few months Disk Utility tells me I have 8 GB free and I spend an evening googling "where does pnpm store its store" and "is it safe to delete DerivedData" (yes). I got tired of it.

cruft is a Go CLI that knows 25 of these paths and cleans them. macOS only, single 5.6 MB binary, ~4,700 LOC.

What it covers: npm / pnpm / yarn / pip / cargo / Go modcache / gem / gradle / maven caches, Xcode DerivedData + archives + simulators, Homebrew downloads, Docker + Colima, .terraform dirs walked from a root, terragrunt cache, pulumi, VS Code + JetBrains caches, Slack, Trash, ~/Library/Caches.

On my M1 it freed 13.2 GB across 19 cleaners in 11 seconds.

How it tries not to wreck your machine:
- Every path is matched against a per-cleaner whitelist. No globs, no "delete everything under X".
- Symlinks are resolved before deletion and rejected if they escape the whitelist root.
- Process-aware: skips the Docker cleaner if Docker.app is running, skips Xcode caches if Xcode is open, etc.
- Risky cleaners (the ones that force a multi-minute re-index or could corrupt a running VM) are opt-in behind --include-risky and each carries its own reason string.
- `--safe` routes deletions through a 7-day recycle bin instead of unlink.
- Every run writes a JSONL audit log.

Two UIs: a bubbletea TUI for picking cleaners interactively, and `cruft run --all` for scripts / cron.

Install (today): `go install github.com/sachincool/cruft/cmd/cruft@latest`
Homebrew tap is queued: `brew install sachincool/tap/cruft`

Repo + README with the full cleaner list and the safety model writeup: https://github.com/sachincool/cruft

What I'd like feedback on:
- Cleaners I'm missing. I don't use Flutter, Conda, Rust-analyzer's project cache, or Bazel daily, so those aren't in yet.
- The risky-cleaner taxonomy: is "may corrupt VM if running" the right level of warning, or do you want a per-cleaner confirm even with --include-risky?
- Linux port. I started this for my own machine. If there's demand I'll factor out the macOS-specific bits (Xcode, Library/Caches, Homebrew prefix detection).

Not on the roadmap: a GUI, a daemon, telemetry, scheduling. It's a CLI you run when the disk is full.
