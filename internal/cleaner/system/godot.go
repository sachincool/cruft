package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "godot",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "Godot editor cache (~/Library/Caches/Godot). Rebuilds on next launch.",
		Paths:         []string{"~/Library/Caches/Godot"},
		DetectAnyPath: true,
		BusyProcs:     []string{"Godot"},
		Reason:        "Godot editor cache",
	})
}
