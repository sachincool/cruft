package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// Symbol/debug caches Xcode builds per device OS version. Regenerated
	// the next time you attach a device on that OS version.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "xcode-devicesupport",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "Xcode device support files (iOS/watchOS/tvOS symbols). Regenerated when you reconnect a device on that OS version.",
		Paths: []string{
			"~/Library/Developer/Xcode/iOS DeviceSupport",
			"~/Library/Developer/Xcode/watchOS DeviceSupport",
			"~/Library/Developer/Xcode/tvOS DeviceSupport",
		},
		DetectAnyPath: true,
		BusyProcs:     []string{"Xcode", "xcodebuild"},
		Reason:        "Xcode device support symbols",
	})
}
