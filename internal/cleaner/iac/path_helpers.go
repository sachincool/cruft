package iac

import "path/filepath"

// pathDir is filepath.Dir wrapped so the helper file is the source of truth.
func pathDir(p string) string  { return filepath.Dir(p) }
func pathBase(p string) string { return filepath.Base(p) }
