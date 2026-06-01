package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// Only the package tarball cache (~/.conda/pkgs). Environments live
	// elsewhere and are never touched here.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "conda",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "Conda's downloaded package tarball cache (~/.conda/pkgs). Re-downloads on next install.",
		Paths:         []string{"~/.conda/pkgs"},
		DetectAnyPath: true,
		BusyProcs:     []string{"conda"},
		Reason:        "conda package cache",
	})
}
