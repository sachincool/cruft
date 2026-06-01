package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "poetry",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "Poetry's artifact + wheel cache. Re-downloads on next install; virtualenvs are untouched.",
		Paths:         []string{"~/Library/Caches/pypoetry"},
		DetectAnyPath: true,
		BusyProcs:     []string{"poetry"},
		Reason:        "poetry download cache",
	})
}
