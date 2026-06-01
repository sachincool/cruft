package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "puppeteer",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "Puppeteer's downloaded Chromium binaries. Re-downloads on next install.",
		Paths:         []string{"~/.cache/puppeteer", "~/Library/Caches/puppeteer"},
		DetectAnyPath: true,
		Reason:        "puppeteer browser cache",
	})
}
