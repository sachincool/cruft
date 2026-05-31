package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "yarn",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "Yarn global cache. Re-downloads on next install.",
		Paths:         []string{"~/Library/Caches/Yarn", "~/.yarn/berry/cache"},
		DetectAnyPath: true,
		BusyProcs:     []string{"yarn", "node"},
		Reason:        "yarn download cache",
	})
}
