package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "deno",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "Deno's dependency cache (remote modules + compiled artifacts). Re-fetches on next run.",
		Paths:         []string{"~/Library/Caches/deno"},
		DetectAnyPath: true,
		BusyProcs:     []string{"deno"},
		Reason:        "deno module cache",
	})
}
