package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "playwright",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "Playwright's downloaded browser binaries. Re-downloads via `playwright install`.",
		Paths:         []string{"~/Library/Caches/ms-playwright"},
		DetectAnyPath: true,
		Reason:        "playwright browser cache",
	})
}
