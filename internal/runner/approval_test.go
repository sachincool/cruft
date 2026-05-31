package runner

import (
	"context"
	"testing"

	"github.com/sachincool/cruft/internal/cleaner"
)

// fakeCleaner is a minimal Cleaner that reports installed, never busy,
// and returns one finding with the configured risk flag.
type fakeCleaner struct {
	name  string
	risky bool
}

func (f fakeCleaner) Name() string                { return f.name }
func (f fakeCleaner) Category() cleaner.Category  { return cleaner.CategorySystem }
func (f fakeCleaner) Description() string         { return "fake" }
func (f fakeCleaner) Risky() bool                 { return f.risky }
func (f fakeCleaner) RiskReason() string          { return "" }
func (f fakeCleaner) Detect(context.Context) bool { return true }
func (f fakeCleaner) BusyProcesses() []string     { return nil }
func (f fakeCleaner) Scan(context.Context, cleaner.ScanOpts) ([]cleaner.Finding, error) {
	return []cleaner.Finding{{Path: "/tmp/" + f.name, Bytes: 1024, Risky: f.risky}}, nil
}
func (f fakeCleaner) Execute(context.Context, []cleaner.Finding, cleaner.ExecOpts) (cleaner.Result, error) {
	return cleaner.Result{}, nil
}

// TestRiskyApproval is the safety invariant for the default (balanced)
// flow: risky findings are scanned and surfaced but stay unapproved, so
// nothing risky executes without the user opting in. AutoApproveRisky
// (set by --include-risky / --profile aggressive) flips that.
func TestRiskyApproval(t *testing.T) {
	cleaners := []cleaner.Cleaner{
		fakeCleaner{name: "safe", risky: false},
		fakeCleaner{name: "danger", risky: true},
	}

	for _, tc := range []struct {
		name             string
		autoApproveRisky bool
		wantRiskyOK      bool
	}{
		{"balanced: risky stays unapproved", false, false},
		{"aggressive: risky auto-approved", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := New(Options{AutoApproveRisky: tc.autoApproveRisky})
			if err != nil {
				t.Fatal(err)
			}
			scans, err := r.Scan(context.Background(), cleaners, nil)
			if err != nil {
				t.Fatal(err)
			}
			for _, s := range scans {
				for _, f := range s.Findings {
					want := true
					if f.Risky {
						want = tc.wantRiskyOK
					}
					if f.Approved != want {
						t.Fatalf("%s finding Approved = %v, want %v", s.Cleaner.Name(), f.Approved, want)
					}
				}
			}
		})
	}
}
