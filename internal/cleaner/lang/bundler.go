package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "bundler",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "Bundler's downloaded gem cache (~/.bundle/cache). Re-downloads on next `bundle install`.",
		Paths:         []string{"~/.bundle/cache"},
		DetectAnyPath: true,
		BusyProcs:     []string{"bundle", "ruby"},
		Reason:        "bundler gem cache",
	})
}
