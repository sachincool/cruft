package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "gem",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "RubyGems download cache. Re-downloads on next `gem install`.",
		Paths:         []string{"~/.gem/ruby"},
		DetectAnyPath: true,
		BusyProcs:     []string{"ruby", "gem", "bundle"},
		Reason:        "gem cache",
	})
}
