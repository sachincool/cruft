# /r/macapps

## Title
[OC] cruft — CLI that reclaims the gigabytes Xcode, Homebrew, and Docker leave on your Mac

## Body
Posting this in case anyone else here writes code on a Mac and has been doing the manual "search for the path, delete the path" dance every time the disk fills up.

`cruft` is a CLI I built for myself. It knows 25 of the places dev tools accumulate gigabytes on macOS and cleans them when you ask. macOS only, single 5.6 MB binary, no GUI, no background process.

Repo: https://github.com/sachincool/cruft

### What it found on my M1
13.2 GB freed across 19 cleaners in 11 seconds.

The biggest offenders, for context:
- **Xcode**: `DerivedData`, old archives, and unused simulator runtimes. If you've ever opened an iOS project this folder is probably multiple GB.
- **Homebrew**: the downloads cache under `~/Library/Caches/Homebrew`. `brew cleanup` covers most of this but doesn't always catch stale downloads.
- **Docker / Colima**: dangling images, build cache, stopped containers. Easy multi-GB win if you haven't run `docker system prune` in a while.

Plus the language tooling that nobody thinks about: npm, pnpm, yarn, pip, cargo, Go modcache, gem, gradle, maven, JetBrains caches, VS Code caches, every `.terraform/` directory under every infra repo you've ever cloned.

### How to use it
Two modes:

- **Interactive TUI** (bubbletea): `cruft` — pick which cleaners to run, see sizes before deleting.
- **Non-interactive**: `cruft run --all` — runs every safe cleaner and exits. Add `--include-risky` if you want the ones that can force re-indexes or affect a running VM.

If you want a safety net, `--safe` routes deletions through a 7-day recycle bin instead of immediate unlink. Every run also writes a JSONL audit log so you can see exactly what got removed.

It refuses to clean a tool's data if that tool is running — opening Docker.app skips the Docker cleaner, Xcode being open skips Xcode caches, etc.

### Disclaimers
- Brand new. Made by one person (me).
- macOS only. No Linux build yet.
- Defaults to direct deletion. If that makes you nervous, add `--safe` for the first run.

### Install
```
go install github.com/sachincool/cruft/cmd/cruft@latest
```
Homebrew tap is queued: `brew install sachincool/tap/cruft`.

Feedback welcome, especially on cleaners I missed.
