package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// Ivy resolution cache + sbt's boot directory (launchers + Scala
	// versions). Both re-download on next build.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "sbt",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "sbt/Ivy resolution cache and boot directory. Re-downloads on next build.",
		Paths:         []string{"~/.ivy2/cache", "~/.sbt/boot"},
		DetectAnyPath: true,
		BusyProcs:     []string{"sbt", "java"},
		Reason:        "sbt/Ivy cache",
	})
}
