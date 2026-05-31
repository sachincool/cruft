package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// We touch only the subdirs that are safe to nuke. ~/.cargo/bin
	// holds user-installed binaries — never delete it.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "cargo",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "Cargo registry + git caches. Next build re-downloads crates.",
		Paths: []string{
			"~/.cargo/registry/cache",
			"~/.cargo/registry/src",
			"~/.cargo/git/checkouts",
			"~/.cargo/git/db",
		},
		DetectAnyPath: true,
		BusyProcs:     []string{"cargo", "rustc"},
		Reason:        "cargo registry/git cache",
	})
}
