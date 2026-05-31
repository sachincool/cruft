package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "jetbrains-caches",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "JetBrains IDE transient caches (~/Library/Caches/JetBrains). Disposable. Project indexes live elsewhere (see jetbrains-system) and are NOT touched.",
		Paths:         []string{"~/Library/Caches/JetBrains"},
		DetectAnyPath: true,
		BusyProcs:     []string{"idea", "pycharm", "rubymine", "phpstorm", "webstorm", "goland", "clion"},
		Reason:        "JetBrains caches",
	})
}
