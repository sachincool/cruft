package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// Windsurf is an Electron/VS Code fork; same cache layout as the
	// vscode cleaner. Extensions and settings are not touched.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "windsurf",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "Windsurf editor caches (Cache, CachedData, GPUCache, logs). Rebuilds on next launch; extensions/settings untouched.",
		Paths: []string{
			"~/Library/Application Support/Windsurf/Cache",
			"~/Library/Application Support/Windsurf/CachedData",
			"~/Library/Application Support/Windsurf/GPUCache",
			"~/Library/Application Support/Windsurf/logs",
		},
		DetectAnyPath: true,
		BusyProcs:     []string{"Windsurf"},
		Reason:        "Windsurf editor cache",
	})
}
