package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// Cursor is an Electron/VS Code fork; same cache layout as the vscode
	// cleaner. Extensions and settings are not touched.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "cursor",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "Cursor editor caches (Cache, CachedData, GPUCache, logs). Rebuilds on next launch; extensions/settings untouched.",
		Paths: []string{
			"~/Library/Application Support/Cursor/Cache",
			"~/Library/Application Support/Cursor/CachedData",
			"~/Library/Application Support/Cursor/GPUCache",
			"~/Library/Application Support/Cursor/logs",
		},
		DetectAnyPath: true,
		BusyProcs:     []string{"Cursor"},
		Reason:        "Cursor editor cache",
	})
}
