package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// Distinct from ~/.gradle/caches (the `gradle` cleaner): this holds
	// the downloaded Gradle distributions the wrapper pins per project.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "gradle-wrapper",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "Gradle wrapper distributions (~/.gradle/wrapper). Re-downloads when a project's wrapper runs.",
		Paths:         []string{"~/.gradle/wrapper"},
		DetectAnyPath: true,
		BusyProcs:     []string{"gradle", "java"},
		Reason:        "gradle wrapper distributions",
	})
}
