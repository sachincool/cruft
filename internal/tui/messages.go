package tui

import "github.com/sachincool/cruft/internal/runner"

// tea.Msg types used by the phase machine.

// scanCompleteMsg signals all scanners have returned.
type scanCompleteMsg struct{ Results []runner.ScanResult }

// execCompleteMsg signals execute phase is done.
type execCompleteMsg struct{ Results []runner.ExecResult }

// errMsg wraps a runtime error so it can be surfaced in the UI.
type errMsg struct{ Err error }

func (e errMsg) Error() string { return e.Err.Error() }
