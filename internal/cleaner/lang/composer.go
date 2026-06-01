package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "composer",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "Composer's download + VCS cache. Re-downloads on next `composer install`.",
		Paths:         []string{"~/.composer/cache", "~/Library/Caches/composer"},
		DetectAnyPath: true,
		BusyProcs:     []string{"composer", "php"},
		Reason:        "composer download cache",
	})
}
