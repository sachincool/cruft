# CRUFT launch plan

## Positioning

CRUFT is the safe cleanup tool for developer laptops: it finds the cache and tool-state bloat that generic cleaners miss, then gives the user dry-run receipts, explicit approvals, audit logs, and an undo window.

Primary promise: **free developer disk space without losing trust.**

## Audience

- macOS developers fighting Xcode, Docker, Homebrew, package-manager, and IaC caches
- platform/dev-experience teams standardizing workstation hygiene
- consultants and power users who rebuild environments often
- open-source contributors who clone and abandon many repos

## Homepage/GitHub checklist

- Add a real TUI screenshot and an asciinema demo.
- Pin a short demo GIF near the top of the README.
- Publish signed release binaries with GoReleaser.
- Create `sachincool/homebrew-tap` and enable the GoReleaser brew section.
- Add repository topics: `macos`, `cli`, `tui`, `disk-cleanup`, `developer-tools`, `xcode`, `docker`, `terraform`, `golang`, `cache-cleaner`.
- Add a short description: `Safe dev-laptop cleanup: dry-run, audit logs, 7-day undo.`
- Turn on GitHub Discussions for cleaner requests and before/after reports.

## Proof users will trust

Ship these before a public launch:

1. tests around every cleaner's owned paths and risky flag
2. fixtures for stale `.terraform/` and `.terragrunt-cache/` detection
3. a documented safety model with diagrams for scan, approve, tombstone, restore
4. real before/after examples with machine-identifying paths redacted
5. a `--json` schema example for scripting

## Promotion plan

- Launch to developer communities with the safety angle, not generic cleanup claims.
- Share concrete examples: Xcode DerivedData, stale Terraform folders, Docker leftovers, Go module cache.
- Ask for cleaner requests with exact paths and regeneration cost.
- Publish a short post: `The caches generic Mac cleaners miss`.
- Record a 60-second demo: scan, explain a cleaner, dry-run JSON, live cleanup, restore.

## Language to use

Use direct, specific wording:

- `dry-run by default`
- `7-day undo`
- `audit log for every action`
- `developer-tool caches, not mystery cleanup`
- `shows paths before touching them`

Avoid vague claims:

- `AI-powered`
- `magic cleanup`
- `optimize your machine`
- `one-click fix everything`
- unverified GB numbers

## Product roadmap

### 0.1 credibility

- CI, release workflow, license, contribution docs
- safety tests for fsutil, tombstone, runner, profile, and each cleaner
- screenshot/demo assets
- Homebrew tap

### 0.2 trust and observability

- `cruft report` with grouped findings and JSON schema docs
- better tombstone listing: size, age, restore preview
- signed checksums or cosign attestations
- cleaner integration tests using temporary fake HOME directories

### 0.3 adoption

- Linux-safe subset
- org policy file for allowed cleaners/profiles
- scheduled reminder mode that only reports
- docs for dev-experience teams
