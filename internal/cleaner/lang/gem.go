package lang

import (
	"path/filepath"

	"github.com/sachincool/cruft/internal/cleaner"
	"github.com/sachincool/cruft/internal/fsutil"
)

func init() {
	// Only the per-version cache/ subdirs hold re-downloadable .gem
	// tarballs. ~/.gem/ruby/<version>/ itself is the user gem home —
	// installed gems, specifications, and executables — which must
	// never be deleted (it would uninstall every user gem).
	paths, _ := filepath.Glob(filepath.Join(fsutil.Expand("~/.gem/ruby"), "*", "cache"))
	paths = append(paths, fsutil.Expand("~/.gem/specs"))
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "gem",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "RubyGems download cache (.gem tarballs and spec cache). Re-downloads on next `gem install`. Installed gems are NOT touched.",
		Paths:         paths,
		DetectAnyPath: true,
		BusyProcs:     []string{"ruby", "gem", "bundle"},
		Reason:        "gem cache",
	})
}
