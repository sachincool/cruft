package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:       "xcode-archives",
		CategoryValue:   cleaner.CategorySystem,
		Desc:            "Xcode .xcarchive bundles created by `Product → Archive`. You may need these for App Store submissions or crash symbolication. Risky by default.",
		IsRisky:         true,
		RiskReasonValue: "may be needed for App Store submissions or crash symbolication",
		Paths:           []string{"~/Library/Developer/Xcode/Archives"},
		DetectAnyPath:   true,
		BusyProcs:       []string{"Xcode", "xcodebuild"},
		Reason:          "Xcode archives (risky — used for App Store / crash symbolication)",
	})
}
