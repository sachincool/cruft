package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "xcode-derived",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "Xcode DerivedData: build intermediates, indexes, module caches. Xcode rebuilds these on next build.",
		Paths:         []string{"~/Library/Developer/Xcode/DerivedData"},
		DetectAnyPath: true,
		BusyProcs:     []string{"Xcode", "xcodebuild"},
		Reason:        "Xcode DerivedData",
	})
}
