package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "prisma",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "Prisma's downloaded query-engine binaries (~/.cache/prisma). Re-downloads on next `prisma generate`.",
		Paths:         []string{"~/.cache/prisma"},
		DetectAnyPath: true,
		Reason:        "prisma engine cache",
	})
}
