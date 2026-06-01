package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "bun",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "Bun's global install cache. Re-downloads on next install; never stores credentials.",
		Paths:         []string{"~/.bun/install/cache"},
		DetectAnyPath: true,
		BusyProcs:     []string{"bun"},
		Reason:        "bun install cache",
	})
}
