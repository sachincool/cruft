package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "vscode",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "VS Code caches: Cache, CachedData, CachedExtensions, GPUCache. Settings, extensions, and project state are NOT touched.",
		Paths: []string{
			"~/Library/Application Support/Code/Cache",
			"~/Library/Application Support/Code/CachedData",
			"~/Library/Application Support/Code/CachedExtensions",
			"~/Library/Application Support/Code/GPUCache",
			"~/Library/Application Support/Code/logs",
		},
		DetectAnyPath: true,
		// VS Code's main process is named "Electron" on macOS, and the
		// helpers are "Code Helper (Renderer)" etc. — the old exact
		// matches "Code"/"Code Helper" never fired, so caches were
		// deleted under a running editor.
		BusyProcs: []string{"Electron", "Code Helper (Renderer)", "Code Helper (GPU)", "Code Helper (Plugin)"},
		Reason:    "VS Code transient caches",
	})
}
