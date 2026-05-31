package iac

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// ~/.pulumi/plugins re-downloads on next `pulumi up`. State is
	// either in your backend or in ~/.pulumi/stacks (NOT here) — safe.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "pulumi",
		CategoryValue: cleaner.CategoryIaC,
		Desc:          "Pulumi plugin cache (~/.pulumi/plugins). State files live elsewhere; this is just downloaded provider plugins.",
		Paths:         []string{"~/.pulumi/plugins"},
		DetectAnyPath: true,
		BusyProcs:     []string{"pulumi"},
		Reason:        "pulumi plugin cache",
	})
}
