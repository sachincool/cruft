package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// "system" contains project indexes — removing them forces a slow
	// re-index on next open. Risky=true.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:       "jetbrains-system",
		CategoryValue:   cleaner.CategorySystem,
		Desc:            "JetBrains 'system' dirs (project indexes). Forces multi-minute re-index on next IDE open. Risky.",
		IsRisky:         true,
		RiskReasonValue: "forces multi-minute project re-index on next IDE open",
		Paths: []string{
			"~/Library/Application Support/JetBrains",
		},
		DetectAnyPath: true,
		BusyProcs:     []string{"idea", "pycharm", "rubymine", "phpstorm", "webstorm", "goland", "clion"},
		Reason:        "JetBrains project indexes (risky — re-index cost)",
	})
}
