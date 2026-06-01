package lang

import (
	"os"

	"github.com/sachincool/cruft/internal/cleaner"
)

func init() {
	// Bazel's default output_user_root is /private/var/tmp/_bazel_<user>;
	// some setups also keep a repository cache under ~/.cache/bazel. Both
	// rebuild on the next `bazel build`.
	paths := []string{"~/.cache/bazel"}
	if u := os.Getenv("USER"); u != "" {
		paths = append(paths, "/private/var/tmp/_bazel_"+u)
	}
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "bazel",
		CategoryValue: cleaner.CategoryLangPkg,
		Desc:          "Bazel output base + repository cache. Rebuilds on next `bazel build` (can be slow).",
		Paths:         paths,
		DetectAnyPath: true,
		BusyProcs:     []string{"bazel"},
		Reason:        "bazel output base + repo cache",
	})
}
