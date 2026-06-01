package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// SDK download cache + downloaded emulator system images. Both
	// re-download via sdkmanager. AVDs (~/.android/avd) hold emulator
	// state/user data and are intentionally NOT touched here.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "android-sdk",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "Android SDK download cache + emulator system images. Re-downloads via sdkmanager. AVD user data is left alone.",
		Paths:         []string{"~/.android/cache", "~/Library/Android/sdk/system-images"},
		DetectAnyPath: true,
		BusyProcs:     []string{"emulator", "adb", "studio"},
		Reason:        "Android SDK cache + system images",
	})
}
