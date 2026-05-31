// Package all blank-imports every cleaner subpackage so their init()
// registration runs. Import this once from cmd/cruft to wire everything.
package all

import (
	_ "github.com/sachincool/cruft/internal/cleaner/container"
	_ "github.com/sachincool/cruft/internal/cleaner/iac"
	_ "github.com/sachincool/cruft/internal/cleaner/lang"
	_ "github.com/sachincool/cruft/internal/cleaner/system"
)
