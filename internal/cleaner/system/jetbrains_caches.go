package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "jetbrains-caches",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "JetBrains IDE caches and project indexes (~/Library/Caches/JetBrains). Regenerate on next IDE open (expect a re-index). Settings, plugins, and licenses in Application Support are NOT touched.",
		Paths:         []string{"~/Library/Caches/JetBrains"},
		DetectAnyPath: true,
		BusyProcs:     []string{"idea", "pycharm", "rubymine", "phpstorm", "webstorm", "goland", "clion"},
		Reason:        "JetBrains caches",
	})
}
