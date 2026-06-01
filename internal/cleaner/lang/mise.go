package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// Only mise's download cache (~/.cache/mise). Installed tool versions
	// live under ~/.local/share/mise and are never touched here.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "mise",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "mise's download cache (~/.cache/mise). Re-downloads when installing a tool; installed versions are untouched.",
		Paths:         []string{"~/.cache/mise"},
		DetectAnyPath: true,
		BusyProcs:     []string{"mise"},
		Reason:        "mise download cache",
	})
}
