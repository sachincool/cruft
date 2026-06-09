package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// SDK download cache only. System images (~/Library/Android/sdk/
	// system-images) are deliberately excluded: every existing AVD
	// breaks until its image is re-downloaded, which is not "junk".
	// AVDs (~/.android/avd) hold emulator state/user data and are
	// intentionally NOT touched either.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "android-sdk",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "Android SDK download cache. Re-downloads via sdkmanager. Emulator system images and AVD user data are left alone.",
		Paths:         []string{"~/.android/cache"},
		DetectAnyPath: true,
		BusyProcs:     []string{"emulator", "adb", "studio"},
		Reason:        "Android SDK download cache",
	})
}
