package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// Only the download cache (~/.cache/pipenv). Virtualenvs live under
	// ~/.local/share/virtualenvs and are never touched here.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "pipenv",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "Pipenv's download cache (~/.cache/pipenv). Re-downloads on next install; virtualenvs are untouched.",
		Paths:         []string{"~/.cache/pipenv"},
		DetectAnyPath: true,
		BusyProcs:     []string{"pipenv"},
		Reason:        "pipenv download cache",
	})
}
