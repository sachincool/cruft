package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "pip",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "pip wheel/HTTP cache. Re-downloads on next install.",
		Paths:         []string{"~/Library/Caches/pip"},
		DetectAnyPath: true,
		BusyProcs:     []string{"pip", "pip3"},
		Reason:        "pip wheel cache",
	})
}
