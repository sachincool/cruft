package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// Xcode's own caches plus the SwiftUI preview cache. Both rebuild on
	// next launch/build. DerivedData and Archives have their own cleaners.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "xcode-caches",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "Xcode app cache + SwiftUI preview cache. Rebuilds on next launch/build.",
		Paths: []string{
			"~/Library/Caches/com.apple.dt.Xcode",
			"~/Library/Developer/Xcode/UserData/Previews",
		},
		DetectAnyPath: true,
		BusyProcs:     []string{"Xcode", "xcodebuild"},
		Reason:        "Xcode app + preview cache",
	})
}
