package lang

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "flutter-pub",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "Dart/Flutter pub package cache + Flutter tool cache. Re-downloads on next `pub get`.",
		Paths:         []string{"~/.pub-cache", "~/Library/Caches/flutter"},
		DetectAnyPath: true,
		BusyProcs:     []string{"dart", "flutter"},
		Reason:        "dart/flutter pub cache",
	})
}
