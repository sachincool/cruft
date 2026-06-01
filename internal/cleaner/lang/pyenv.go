package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// Only pyenv's build cache (~/.pyenv/cache). Installed Python versions
	// live under ~/.pyenv/versions and are never touched here.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "pyenv",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "pyenv's source download cache (~/.pyenv/cache). Re-downloads when building a version; installed versions are untouched.",
		Paths:         []string{"~/.pyenv/cache"},
		DetectAnyPath: true,
		Reason:        "pyenv download cache",
	})
}
