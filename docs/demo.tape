# VHS script for the README hero GIF.
#
# Install once:  brew install vhs
# Render:        vhs docs/demo.tape   ->   assets/demo.gif
#
# Hold the recording to ~25-30s. Three beats:
#   1. cruft doctor   — establishes scope, builds trust
#   2. cruft          — the TUI moment (scan → pick → confirm)
#   3. cruft glossary — shows the tool teaches itself
#
# Everything is --dry-run so nothing on the recording machine actually
# gets deleted. The summary still shows reclaimable totals.

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

# Quiet shell prompt for the recording — single chevron, no clutter.
Hide
Type "export PS1='%F{green}❯%f '"
Enter
Type "clear"
Enter
Show

# --- Beat 1: doctor (what's installed, what's busy, what's risky) ---
Type "cruft doctor"
Sleep 400ms
Enter
Sleep 4s

Type "clear"
Enter
Sleep 200ms

# --- Beat 2: the TUI ---
Type "cruft --dry-run"
Sleep 400ms
Enter

# Scan animation
Sleep 3500ms

# Browse the list briefly so the viewer sees the cursor + sizes.
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 600ms

# Open the help panel on the focused cleaner.
Type "?"
Sleep 2200ms

# Dismiss help, accept the default selection.
Space
Sleep 600ms
Enter
Sleep 1500ms

# Confirm screen: hit y, watch the dry-run "execute" complete instantly.
Type "y"
Sleep 2200ms

# Summary screen sits long enough for the viewer to read totals.
Sleep 3000ms

# Quit.
Type "q"
Sleep 600ms

Type "clear"
Enter
Sleep 200ms

# --- Beat 3: cruft teaches itself ---
Type "cruft glossary safe-mode"
Sleep 400ms
Enter
Sleep 3500ms

# End frame.
Type "clear"
Enter
Sleep 800ms
