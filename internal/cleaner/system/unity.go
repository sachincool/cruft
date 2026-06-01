package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// Downloaded Asset Store packages + the shared GI/package cache.
	// Asset Store packages re-download from your account; the cache
	// rebuilds on next import/bake.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "unity",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "Unity Asset Store downloads + shared GI/package cache. Re-downloads / rebuilds on next use.",
		Paths: []string{
			"~/Library/Unity/Asset Store-5.x",
			"~/Library/Unity/cache",
			"~/Library/Caches/com.unity3d.UnityEditor",
		},
		DetectAnyPath: true,
		BusyProcs:     []string{"Unity", "Unity Hub"},
		Reason:        "Unity asset + GI cache",
	})
}
