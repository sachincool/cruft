package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "swiftpm",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "Swift Package Manager's downloaded packages. Re-downloads on next build.",
		Paths:         []string{"~/Library/Caches/org.swift.swiftpm", "~/.swiftpm"},
		DetectAnyPath: true,
		BusyProcs:     []string{"swift", "xcodebuild"},
		Reason:        "Swift PM package cache",
	})
}
