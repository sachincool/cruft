package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// Only nvm's download cache (~/.nvm/.cache). Installed Node versions
	// live under ~/.nvm/versions and are never touched here.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "nvm",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "nvm's Node download cache (~/.nvm/.cache). Re-downloads when installing a version; installed versions are untouched.",
		Paths:         []string{"~/.nvm/.cache"},
		DetectAnyPath: true,
		Reason:        "nvm download cache",
	})
}
