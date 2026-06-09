// Package profile defines preset selection levels (conservative,
// balanced, aggressive) that filter the registered cleaners.
package profile

import (
	"strings"

	"github.com/sachincool/cruft/internal/cleaner"
)

type Profile string

const (
	Conservative Profile = "conservative"
	Balanced     Profile = "balanced"
	Aggressive   Profile = "aggressive"
)

// Parse returns the named profile (case-insensitively), or Balanced if
// name is empty / unrecognised. Callers that want to reject typos
// instead of silently degrading should check Valid first.
func Parse(name string) Profile {
	switch p := Profile(strings.ToLower(name)); p {
	case Conservative, Aggressive:
		return p
	}
	return Balanced
}

// Valid reports whether name is a recognised profile. "" is valid (it
// means "use the default").
func Valid(name string) bool {
	switch Profile(strings.ToLower(name)) {
	case "", Conservative, Balanced, Aggressive:
		return true
	}
	return false
}

// fastRegenerate names cleaners whose post-cleanup cost is genuinely
// negligible: just re-download or lazy re-derive on next use, no
// project rebuild and no multi-minute pause.
//
// Conservative previously included xcode-derived; that was wrong —
// wiping DerivedData typically triggers a 5–15 minute rebuild on the
// next compile of any non-trivial Xcode project. Removed.
//
// Curated, not derived.
var fastRegenerate = map[string]bool{
	"npm":              true,
	"pnpm":             true,
	"yarn":             true,
	"pip":              true,
	"gem":              true,
	"homebrew":         true,
	"xcode-simulators": true, // delete unavailable only; no rebuild cost
	"vscode":           true,
	"jetbrains-caches": true, // caches + indexes; regenerate on next IDE open
	"slack":            true,
	// pure re-download caches with no project rebuild
	"bun":      true,
	"deno":     true,
	"uv":       true,
	"poetry":   true,
	"pipenv":   true,
	"conda":    true,
	"composer": true,
	"bundler":  true,
	"nvm":      true,
	"pyenv":    true,
	"mise":     true,
	// rebuild-on-next-launch app/tool caches
	"cursor":              true,
	"windsurf":            true,
	"godot":               true,
	"aws-cli":             true,
	"xcode-caches":        true,
	"xcode-devicesupport": true,
	// re-download browser/engine binaries, no project rebuild
	"playwright": true,
	"puppeteer":  true,
	"prisma":     true,
}

// Filter returns the subset of cleaners that match the profile.
//
//   - Conservative: only fast-regenerate caches (no risky). Tight scope.
//   - Balanced: everything is scanned, including risky. Runner auto-
//     disapproves risky findings — the user can opt in via the TUI or
//     by re-running with --include-risky. Risky stays *visible* so the
//     user knows what's available without having to guess.
//   - Aggressive: everything; runner auto-approves risky too.
//
// Auto-approval of risky findings is decided by the runner (see
// runner.Options.AutoApproveRisky), not here. Filter just chooses
// which cleaners to scan at all.
func Filter(all []cleaner.Cleaner, p Profile) []cleaner.Cleaner {
	out := make([]cleaner.Cleaner, 0, len(all))
	for _, c := range all {
		if p == Conservative {
			if c.Risky() {
				continue
			}
			if !fastRegenerate[c.Name()] {
				continue
			}
		}
		// Balanced & Aggressive: include everything; auto-approval
		// behaviour differs at runner level.
		out = append(out, c)
	}
	return out
}
