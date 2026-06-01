package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "uv",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "uv's wheel + source cache. Re-downloads on next install; never stores credentials.",
		Paths:         []string{"~/.cache/uv", "~/Library/Caches/uv"},
		DetectAnyPath: true,
		BusyProcs:     []string{"uv"},
		Reason:        "uv download cache",
	})
}
