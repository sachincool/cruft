# Product Hunt

## Tagline (≤60 chars)
The CLI that knows where your Mac hides 13 GB of dev junk

## Description (≤260 chars)
cruft is a macOS CLI with 25 cleaners for the caches dev tools leave behind — Xcode DerivedData, Homebrew, Docker, npm/pnpm/cargo/pip, .terraform dirs, JetBrains, the lot. Interactive TUI or `cruft run --all`. Whitelist-only deletes, audit log per run. 5.6 MB Go binary.

---

## First comment from the maker

Hi PH — I'm Sachin. I built cruft because every two months my 256 GB MacBook fills up and I spend an evening googling "where does pnpm put its store" and "is it safe to delete DerivedData".

The honest pitch: there are 25 cleaners. They cover the language caches (npm, pnpm, yarn, pip, cargo, Go modcache, gem, gradle, maven), the IaC state dirs that pile up if you do any terraform/terragrunt/pulumi work, Docker and Colima, and the macOS-specific ones — Xcode DerivedData and archives and simulators, Homebrew, JetBrains, VS Code, Slack, Trash, `~/Library/Caches`. On my M1 it freed 13.2 GB in 11 seconds.

The bit I actually care about is the safety model. Every cleaner has a whitelist of target paths — there are no glob patterns and no "rm under this dir". Symlinks get resolved before deletion and rejected if they point outside the whitelist root. The cleaners that could corrupt a running VM or force a long re-index are opt-in behind `--include-risky`, and each one prints its own reason at runtime. If you want a safety net, `--safe` routes everything through a 7-day recycle bin.

Caveats I want to be upfront about: it's macOS only, it's brand new, and it's written by one person. Not on the roadmap: a GUI, a daemon, telemetry, scheduling. It's a CLI you run when the disk is full. Repo and full cleaner list: https://github.com/sachincool/cruft
