package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// Allowlist of ~/Library/Caches subdirs known to be regeneratable.
	// We never blanket-delete the parent — some apps stash auth/license
	// tokens there.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "library-caches",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "Vetted subdirs under ~/Library/Caches that are known to regenerate cleanly. Allowlist-only — apps that stash auth/license data are never touched.",
		Paths: []string{
			"~/Library/Caches/Google/Chrome",
			"~/Library/Caches/com.apple.Safari",
			"~/Library/Caches/Firefox",
			"~/Library/Caches/Homebrew",
			"~/Library/Caches/com.spotify.client",
			"~/Library/Caches/com.tinyspeck.slackmacgap",
		},
		DetectAnyPath: true,
		// Deleting a live Chromium/Gecko cache directory is a known
		// corruption vector — skip while any of these apps run, same as
		// xcode-derived does for Xcode.
		BusyProcs: []string{"Google Chrome", "Safari", "firefox", "Spotify", "Slack"},
		Reason:    "vetted system caches",
	})
}
