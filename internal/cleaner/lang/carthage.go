package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "carthage",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "Carthage's downloaded dependency cache. Re-fetches on next `carthage bootstrap`.",
		Paths:         []string{"~/Library/Caches/org.carthage.CarthageKit"},
		DetectAnyPath: true,
		BusyProcs:     []string{"carthage"},
		Reason:        "carthage download cache",
	})
}
