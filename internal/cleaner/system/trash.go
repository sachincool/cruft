package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:       "trash",
		CategoryValue:   cleaner.CategorySystem,
		Desc:            "Empty ~/.Trash. Per-volume trashes on external drives are NOT touched. Files in Trash are unrecoverable after this (use the tombstone or recover them from Finder first if unsure).",
		IsRisky:         true,
		RiskReasonValue: "deleted files are unrecoverable unless tombstone is on",
		Paths:           []string{"~/.Trash"},
		DetectAnyPath:   true,
		Reason:          "Trash (risky — unrecoverable after delete unless tombstoned)",
	})
}
