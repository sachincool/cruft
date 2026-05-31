# Twitter / X thread

---

**Tweet 1/5**

shipped a thing: cruft

a macOS CLI that knows where your dev tools hide their gigabytes and deletes them on request

ran it on my M1 last night, got 13.2 GB back in 11 seconds across 19 cleaners

[screenshot: TUI showing the cleaner list with sizes next to each row]

---

**Tweet 2/5**

25 cleaners total. the usual suspects (npm, pnpm, yarn, pip, cargo, Go modcache, gem, gradle, maven) plus the ones that actually eat the disk:

~/Library/Developer/Xcode/DerivedData
Xcode archives + simulators
Homebrew downloads
Docker + Colima VMs

---

**Tweet 3/5**

the safety model is the part i care about:

- every cleaner has a whitelist of paths. no globs.
- symlinks resolved before delete, rejected if they escape the whitelist root
- skips Docker if Docker is running, Xcode if Xcode is open, etc.
- risky cleaners opt-in behind --include-risky

---

**Tweet 4/5**

two modes:

interactive bubbletea TUI for browsing what's about to die

`cruft run --all` for scripts. add `--safe` to route through a 7-day recycle bin instead of unlink. every run writes a JSONL audit log so you can grep what got deleted.

---

**Tweet 5/5**

single 5.6 MB Go binary, macOS only for now, brand new

`go install github.com/sachincool/cruft/cmd/cruft@latest`

repo: https://github.com/sachincool/cruft

tell me which cleaner i forgot
