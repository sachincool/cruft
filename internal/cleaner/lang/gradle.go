package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "gradle",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "Gradle build cache and downloaded dependencies. Repopulates on next build.",
		Paths:         []string{"~/.gradle/caches"},
		DetectAnyPath: true,
		BusyProcs:     []string{"gradle", "java"},
		Reason:        "gradle cache",
	})
}
