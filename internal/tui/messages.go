package tui

import "github.com/sachincool/cruft/internal/runner"

// tea.Msg types used by the phase machine.

// scanProgressMsg carries one cleaner's result as soon as it finishes,
// so the scan view can tick up live instead of waiting for the whole batch.
type scanProgressMsg struct{ Result runner.ScanResult }

// scanCompleteMsg signals every scanner has returned (the progress
// channel closed). Results are already accumulated on the model.
type scanCompleteMsg struct{}

// execCompleteMsg signals execute phase is done. SnapReclaimed is the
// free space (bytes) recovered by thinning Time Machine local snapshots
// after a live run; 0 when none / dry-run.
type execCompleteMsg struct {
	Results       []runner.ExecResult
	SnapReclaimed int64
}

// errMsg wraps a runtime error so it can be surfaced in the UI.
type errMsg struct{ Err error }

func (e errMsg) Error() string { return e.Err.Error() }
