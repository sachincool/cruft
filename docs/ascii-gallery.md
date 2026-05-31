# cruft — ASCII gallery

Source-of-truth for the wordmark, hero block, mock TUI screenshots, and
dividers used in `README.md` and the CLI banner. Every block is
hand-aligned for monospace; copy-paste verbatim.

---

## 1. Wordmark (three variants)

### Variant 1 — modern half-block (recommended)

```text
█▀▀ █▀▀▄ █  █ █▀▀ ▀█▀   ░▒▓
█   █▀▀▄ █  █ █▀▀  █     ░▒▓
▀▀▀ ▀  ▀ ▀▀▀▀ ▀    ▀      ░▒
```

Three-row half-block letterforms with a stepped fade ramp on the right —
each row drops one shade, so the eye reads the letters being "eaten" from
the top-right corner. Tight 21-col body, 27 cols overall.

### Variant 2 — four-row solid (impact)

```text
█▀▀▀ █▀▀▄ █   █ █▀▀▀ ▀▀█▀▀  ▓
█    █▀▀▄ █   █ █▀▀    █    ▓▒
█    █  █ █   █ █      █    ▓▒░
▀▀▀▀ ▀  ▀  ▀▀▀  ▀      ▀    ▒░
```

Heavier four-row build. Fade is a triangular wedge that grows downward,
so the bottom right is most eaten. Reads as a banner — use this for the
CLI `--version` splash and the top of the README, not in body copy.

### Variant 3 — experimental notched-T

```text
█▀▀▖ █▀▀▄ █  █ █▀▀ ▀█▀   ▓▒░
█    █▀▀▄ █  █ █▀▀  █     ▒░
▀▀▀▘ ▀  ▀ ▀▀▀▀ ▀    ▀      ░
```

Same skeleton as Variant 1 but with quadrant corners (`▖▘`) carved into
the C, treating the whole wordmark as one cut shape rather than five
discrete glyphs. More opinionated — argue for it if cruft is doubling
down on a brutalist mark; argue against if Variant 1 has to survive on
every terminal font.

---

## 2. README hero block

```text
┌──────────────────────────────────────────┐   █▀▀ █▀▀▄ █  █ █▀▀ ▀█▀  ░▒▓
│ before   ██████████████████████   47 GB  │   █   █▀▀▄ █  █ █▀▀  █    ░▒
│  after   █████▓▒░░░░░░░░░░░░░░░   12 GB  │   ▀▀▀ ▀  ▀ ▀▀▀▀ ▀    ▀     ░
└──────────────────────────────────────────┘   decruft your dev laptop.
                  −35 GB freed                 one binary · 25 cleaners · macOS
```

Width: 79 cols. Height: 5 rows. The before/after bar is the visual hook —
the "after" row literally uses the brand fade (`█▓▒░`) to show what got
eaten. The `−35 GB freed` sits centered under the box, accent-green in
the rendered README.

---

## 3. Mock TUI screenshots

### 3a. `cruft doctor`

```text
$ cruft doctor

Legend  ready = safe to clean  ·  in use = skipped (a process is running)
        risky = unchecked by default; run `cruft explain <name>` for why

  pnpm                    ready
  npm                     in use (node)
  yarn                    not installed
  cargo                   ready
  gomod                   ready
  pip                     ready
  gem                     not installed
  gradle                  in use (java)
  maven                   ready risky · re-downloads every Maven dep on next build
  pulumi                  ready
  terraform               ready
  terragrunt              not installed
  docker                  ready
  docker-volumes          ready risky · anonymous volumes often hold db/state
  colima                  in use (colima)
  homebrew                ready
  library-caches          ready
  vscode                  in use (Code)
  jetbrains-caches        ready
  jetbrains-system        ready risky · forces multi-minute project re-index
  xcode-derived           ready
  xcode-archives          ready risky · may be needed for App Store submissions
  xcode-simulators        ready
  slack                   in use (Slack)
  trash                   ready risky · unrecoverable unless tombstone is on

Search roots:
  /Users/you
  /Users/you/code

Tombstone size: 0 B
Tip: `cruft glossary <term>` defines any word here · `cruft explain <name>`
```

### 3b. TUI select screen (mid-session)

```text
  C R U F T ░▒▓
  decruft your laptop

What to clean
space toggle · a toggle-all · ? what is this · d preview · s safe-mode · enter confirm · q quit

  [d] preview off   [s] safe-mode ON (7d recycle)

LANGUAGE / PACKAGE
  [✓] pnpm                 4.1 GB     /Users/you/Library/pnpm/store
▶ [✓] cargo                2.7 GB     target/ across 9 repos
  [✓] gomod                1.8 GB     GOMODCACHE
  [✓] pip                  612 MB     ~/Library/Caches/pip
  [ ] maven                430 MB !   re-downloads every Maven dep on next build

CONTAINER
  [✓] docker               6.2 GB     images, build cache, stopped containers
  [ ] docker-volumes       1.1 GB !   anonymous volumes often hold db/state

SYSTEM
  [✓] xcode-derived        3.9 GB     DerivedData
  [✓] xcode-simulators     2.4 GB     14 unavailable runtimes
  [✓] homebrew             880 MB     brew cleanup -s
  [✓] library-caches       310 MB     6 vetted subdirs
  [ ] xcode-archives       2.1 GB !   used for App Store / crash symbolication
  [ ] trash                740 MB !   unrecoverable after delete unless tombstoned

·  3 not installed   ·  2 already clean   ·  4 skipped (in use)

──────────────────────────────────────────────────────────────────────
selected: 22.9 GB across 9 cleaners                       enter confirm
```

### 3c. Summary after a real run

```text
cruft summary  (live)

  ✓ pnpm                   4.1 GB
  ✓ cargo                  2.7 GB
  ✓ gomod                  1.8 GB
  ✓ pip                    612 MB
  ✓ docker                 6.2 GB
  ✓ xcode-derived          3.9 GB
  ✓ xcode-simulators       2.4 GB
  ✓ homebrew               880 MB
  ✓ library-caches         310 MB

⚠  3.9 GB more available from risky cleaners (not executed):
    xcode-archives         2.1 GB   · may be needed for App Store submissions
    docker-volumes         1.1 GB   · anonymous volumes often hold db/state
    trash                  740 MB   · unrecoverable after delete unless tombstoned
    to also clean these, re-run with `--include-risky` or `--profile aggressive`

⏸  skipped because a related tool is running:
    npm (node in use)
    gradle (java in use)
    colima (colima in use)
    vscode (Code in use)
    slack (Slack in use)
    re-run after the tool exits to capture this reclaim.

not installed on this machine: yarn, gem, terragrunt

Total: 22.9 GB
Audit log: ~/.local/share/cruft/audit/2026-05-31T08-42-17Z.jsonl
Tombstone: ~/.local/share/cruft/tombstone  (restore with `cruft restore 20260531-084217`)

What's next
  cruft restore 20260531-084217   undo this run (within 7 days)
  cruft last                      per-cleaner detail for this run
  cruft history                   all past runs
  cruft                           run again
```

---

## 4. Section dividers

### Variant 1 — minimal

```text
────────────────────────────────────────────────────────────────────────────────
```

A single horizontal rule, full 80 cols. Use between body sections where
the heading already carries the weight.

### Variant 2 — brand fade

```text
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ ░▒▓
```

Heavy rule with the brand fade tucked at the right edge. Use once per
page — typically right under the hero block — to echo the wordmark.

### Variant 3 — shaded ramp

```text
░░░░░▒▒▒▒▒▓▓▓▓▓██████████████████████████████████████▓▓▓▓▓▒▒▒▒▒░░░░░
```

A symmetric ramp that builds to solid and fades back. Use as a section
break between major chapters (Install / Use / Safety / Develop) when you
want a visible page-turn.
