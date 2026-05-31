package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// We don't blanket-delete ~/.m2/repository — that holds your
	// declared dependencies and re-downloads can be slow on flaky
	// networks. Only the local install cache for snapshots is safe.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:       "maven",
		CategoryValue:   cleaner.CategoryLangPkg,
		Desc:            "Maven repository cache (~/.m2/repository). Re-downloads on next build.",
		IsRisky:         true,
		RiskReasonValue: "re-downloads every Maven dep on next build (slow on poor networks)",
		Paths:           []string{"~/.m2/repository"},
		DetectAnyPath:   true,
		BusyProcs:       []string{"mvn", "java"},
		Reason:          "maven repo cache (risky — re-downloads all deps)",
	})
}
