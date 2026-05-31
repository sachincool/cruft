package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "npm",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "npm package cache. Re-downloads on next install; never stores credentials.",
		Paths:         []string{"~/.npm/_cacache"},
		DetectAnyPath: true,
		BusyProcs:     []string{"npm", "node"},
		Reason:        "npm download cache",
	})
}
