package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "pnpm",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "pnpm content-addressable store. Repopulates on next install.",
		Paths:         []string{"~/Library/pnpm/store", "~/.local/share/pnpm/store"},
		DetectAnyPath: true,
		BusyProcs:     []string{"pnpm", "node"},
		Reason:        "pnpm store cache",
	})
}
