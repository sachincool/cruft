# I got tired of googling where pnpm hides its cache, so I wrote cruft

It was 1 a.m. on a Tuesday. I'd been trying to `git clone` a 400 MB monorepo and the clone kept dying halfway through with "no space left on device". Disk Utility said I had 3 GB free on a 256 GB disk. I knew, in the abstract, that most of the missing space was dev caches. I had no idea, in the specific, where any of them lived.

So I opened a new tab and typed "where does pnpm store its global store macos". Then "is it safe to delete xcode derived data". Then "homebrew clear cache command". Then "find all .terraform folders". Forty minutes later I'd freed about 18 GB by hand and I was annoyed enough that I'd stopped being sleepy. The annoyance was the useful signal: I'd done this exact dance four or five times in the last year, and I was going to do it again in two months, and the next person sitting at a Mac with a full disk at 1 a.m. wasn't going to have a better time than I'd just had.

The fix is obvious in retrospect: a CLI that knows the paths so I don't have to remember them. I called it `cruft`.

## What it does

`cruft` ships 25 cleaners. Each cleaner knows about one tool and where that tool puts data you can safely delete. The list, as of today:

- **Language package managers**: npm, pnpm, yarn, pip, cargo, Go modcache, gem, gradle, maven
- **IaC**: terraform (it walks from a root and finds every `.terraform/` under it), terragrunt cache, pulumi
- **Containers**: Docker, Colima
- **Mac apps and system**: Homebrew downloads, Xcode DerivedData, Xcode archives, Xcode simulators, VS Code caches, JetBrains caches and system dirs, Slack, Trash, `~/Library/Caches`

You can run it two ways. The first is the TUI:

```
$ cruft
```

This opens a bubbletea interface that lists every cleaner, the paths it targets, and the estimated bytes it would free. You arrow through, hit space to toggle, and enter to run.

The second is for scripts and for people who already know what they want:

```
$ cruft run --all
```

On my M1, that command freed **13.2 GB across 19 cleaners in 11 seconds** the first time I ran it on a real machine. The cleaners it skipped were either ones with nothing to clean (cold cargo cache) or ones it refused to run because the underlying tool was open.

## The design philosophy, such as it is

I'll be honest, "design philosophy" is generous for a tool this small. But the deletion model has four rules I worked out before I wrote any cleaners, and they shaped everything:

### 1. Whitelist-only path matching

Every cleaner declares the exact paths it's allowed to touch. There are no glob patterns. There is no `find ... -exec rm`. The Xcode DerivedData cleaner can delete things under `~/Library/Developer/Xcode/DerivedData` and nowhere else. If you write a new cleaner and you forget to add its root to the whitelist, the cleaner does nothing — it doesn't fall back to "well, try anyway".

This sounds restrictive and it is. That's the point. Most data-loss disasters I've seen in shell-script land come from a globbing edge case or a `$VAR` that turned out to be empty so `rm -rf $VAR/foo` became `rm -rf /foo`. With a whitelist, the worst a buggy cleaner can do is delete things inside its declared root, which is a much smaller blast radius.

### 2. Symlinks get resolved, escapes get rejected

Before any file gets deleted, its path is run through `filepath.EvalSymlinks`. If the resolved path lives outside the cleaner's whitelist root, the entry is skipped and logged. So if some package manager decided to symlink its cache into `~/important-work`, cruft won't follow the link out and start deleting your repo. It'll just skip that entry.

I tested this with deliberately constructed evil trees. It holds up.

### 3. Risky cleaners are opt-in, with reasons

Not every cleaner is equally safe to run. Deleting Homebrew's downloads cache costs you a re-download next time you `brew install`. Deleting Xcode's module cache forces a multi-minute re-index. Deleting a Colima VM's overlay while the VM is running might corrupt the VM.

These three things should not have the same default. So cruft splits cleaners into two tiers. The default tier runs on `cruft run --all`. The risky tier requires `--include-risky`. Every risky cleaner carries a string explaining *why* it's risky, and the string prints when the cleaner runs. Things like:

> Xcode module cache: forces a multi-minute re-index on next build.

> Colima overlay: may corrupt the VM if Colima is running. Stop Colima first.

You don't get a vague "this is risky, are you sure?" prompt. You get the actual mechanism.

The risky tier also composes with process awareness: even with `--include-risky`, cruft won't clean Docker if Docker.app is running, won't clean Xcode caches if Xcode is open, etc. The flag is "I accept the cost", not "skip safety checks".

### 4. The recycle bin is opt-in, not the default

This is the most opinionated bit and I'll defend it.

Lots of cleanup tools route deletions through a recycle bin or a trash folder by default. The argument is "what if you wanted that file back?". My counter-argument is: if you're running a tool called `cruft` against your `~/Library/Caches`, you have already decided you don't want those bytes. Wrapping them in a recycle layer just doubles the disk usage temporarily and gives a false sense of safety.

So the default is direct deletion. If you want the safety net, `--safe` puts everything through a 7-day recycle bin under cruft's cache directory. I use `--safe` the first time I run cruft on a new machine and never again. Your mileage may vary.

### 5. (Bonus) Every run writes an audit log

I lied, there are five rules. The last one is that every run appends a JSONL audit log: cleaner name, path, bytes freed, error if any. So when you wake up the next morning wondering "wait, did cruft delete my Postman scratch files", you can `jq` the log and find out.

## A walkthrough of a real session

Here's roughly what last week's cleanup looked like.

```
$ cruft
```

The TUI opened. The top of the list was `xcode-derived-data` at 4.8 GB, which surprised me until I remembered I'd been bouncing between three iOS branches this month. Below that, `homebrew-downloads` at 2.1 GB, `docker` at 1.9 GB, `pnpm-store` at 1.4 GB, `cargo-cache` at 800 MB, then a long tail of cleaners under 500 MB each.

I hit `a` to select all, then enter. cruft showed a confirm screen with the total — 13.2 GB — and the list of skipped cleaners with reasons:

```
skipped: docker (Docker.app is running)
skipped: vscode-caches (Code Helper process running)
```

I closed Docker and VS Code, hit `r` to re-scan, confirmed, and watched the bar fill up. Eleven seconds later it was done.

Then I checked the audit log out of curiosity:

```
$ tail -3 ~/Library/Application\ Support/cruft/audit.jsonl | jq .
{"cleaner":"xcode-derived-data","path":"/Users/me/Library/Developer/Xcode/DerivedData","bytes":4831234567,"err":null}
{"cleaner":"homebrew-downloads","path":"/Users/me/Library/Caches/Homebrew/downloads","bytes":2110293847,"err":null}
{"cleaner":"docker","path":"docker system prune -a -f","bytes":1923847562,"err":null}
```

That's the entire UX. There's nothing else.

## What it doesn't do

Worth being clear about the limits.

**It doesn't run on Linux.** A bunch of cleaners are specifically about macOS paths — `~/Library/Caches`, `~/Library/Developer/Xcode/*`, Homebrew's Mac layout — and untangling those from the cross-platform cleaners is a refactor I haven't done. If there's interest I'll do it.

**It doesn't have a GUI.** I don't want one. The TUI is enough for interactive use and `cruft run --all` is enough for scripts. A GUI would be a second surface to maintain.

**It doesn't run in the background.** No daemon, no scheduled scans, no launchd job. It's a tool you run when the disk is full. If you want it scheduled, write a cron entry.

**It doesn't send any telemetry.** It doesn't phone home, it doesn't update itself, it doesn't know who you are. The audit log lives on your machine and only your machine.

**It doesn't cover everything.** Things I don't use daily and therefore haven't written cleaners for: Flutter's pub cache, Conda envs, Bazel's output base, Rust-analyzer's project cache, Android Studio's intermediate dirs. The cleaner interface is small enough that adding one is maybe 50 lines, so if there's a tool you'd like covered, file an issue or send a PR.

**It doesn't promise to never delete something you wanted.** It tries hard — whitelists, symlink resolution, process checks, opt-in risky tier, opt-in recycle bin, audit log — but it is a deletion tool written by a single human and you should treat it as such. Use `--safe` on the first run.

## What's next

In rough order:

- **Homebrew tap**, so `brew install sachincool/tap/cruft` works. This is queued.
- **Linux support**, if there's interest. Probably easier to start a `cruft-linux` build than to maintain a unified binary, but I haven't decided.
- **A few more cleaners** — pub, conda, bazel are the top of the list.
- **A `--dry-run` that's more honest** than scanning-then-asking. Right now you get sizes in the TUI before confirming, but `cruft run --all` has no equivalent. It should.
- **Output formats** for the JSONL audit log: maybe a `cruft log --since 7d` that pretty-prints recent activity.

That's it. cruft is small, it does one thing, and the thing it does is "remember where dev tools hide gigabytes on a Mac so you don't have to at 1 a.m."

Repo, README with the full cleaner list, and the safety-model writeup: https://github.com/sachincool/cruft

Install today:

```
go install github.com/sachincool/cruft/cmd/cruft@latest
```

If you run it and it frees a comically large number of gigabytes, I'd love to hear what the worst offender was.
