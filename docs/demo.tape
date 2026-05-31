# VHS script for the README hero GIF.
#
# Install once:  brew install vhs
# Render:        vhs docs/demo.tape   ->   assets/demo.gif
#
# ~18s, three beats:
#   1. cruft doctor                    — what's installed / busy / risky
#   2. cruft gomod pip library-caches  — the TUI: scan, review, confirm
#   3. cruft glossary safe-mode        — the tool documents itself
#
# Beat 2 names three cleaners so the scan is instant on any machine —
# a bare `cruft` scans all 25, and shell-outs like `brew cleanup` make
# that take 15s+, which is dead air in a recording. Everything is
# --dry-run, so nothing is deleted while recording.

Output assets/demo.gif

Set Theme "Catppuccin Mocha"
Set FontSize 16
Set Width 1180
Set Height 720
Set TypingSpeed 32ms
Set PlaybackSpeed 1.0
Set Padding 28
Set Margin 12
Set MarginFill "#1a1a1a"
Set Framerate 30

# Clean single-chevron bash prompt. Set while hidden so the PATH line and
# any shell-startup noise never make it into the recording.
Hide
Type "export PATH=$HOME/go/bin:$PATH"
Enter
Type `export PS1='\[\e[1;32m\]❯\[\e[0m\] '`
Enter
Type "clear"
Enter
Show

# --- Beat 1: doctor (what's installed, what's busy, what's risky) ---
Type "cruft doctor"
Sleep 400ms
Enter
Sleep 2500ms

Type "clear"
Enter
Sleep 200ms

# --- Beat 2: the TUI, scoped to three cleaners for a snappy scan ---
Type "cruft gomod pip library-caches --dry-run"
Sleep 400ms
Enter

# Scan resolves to the select screen.
Sleep 2200ms

# Move the cursor so the viewer sees navigation and per-cleaner sizes.
Down
Sleep 400ms
Down
Sleep 600ms

# Open the help panel on the focused cleaner, then dismiss it.
Type "?"
Sleep 2000ms
Escape
Sleep 500ms

# Confirm, then watch the dry-run execute land on the summary.
Enter
Sleep 700ms
Type "y"
Sleep 2600ms

# Quit from the summary.
Type "q"
Sleep 500ms

Type "clear"
Enter
Sleep 200ms

# --- Beat 3: cruft documents itself ---
Type "cruft glossary safe-mode"
Sleep 400ms
Enter
Sleep 2800ms

Type "clear"
Enter
Sleep 600ms
