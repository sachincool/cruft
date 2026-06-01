package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// The spec repo (~/.cocoapods/repos) can be several GB; it re-clones
	// on `pod install`/`pod repo update`. The download cache rebuilds too.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "cocoapods",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "CocoaPods spec repo + download cache. Re-clones on next `pod install`.",
		Paths:         []string{"~/Library/Caches/CocoaPods", "~/.cocoapods/repos"},
		DetectAnyPath: true,
		BusyProcs:     []string{"pod"},
		Reason:        "cocoapods spec repo + cache",
	})
}
