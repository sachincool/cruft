package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sachincool/cruft/internal/cleaner"
)

// RunSummary aggregates one past run from its JSONL log.
type RunSummary struct {
	RunID       string
	Path        string
	Entries     int
	Successes   int
	Failures    int
	BytesFreed  int64
	DryRun      bool
	PerCleaner  map[string]int64
	FirstAction string // RFC3339 timestamp of first entry
	LastAction  string
}

// History reads all *.jsonl files in dir and returns one RunSummary
// per run, newest first.
func History(dir string) ([]RunSummary, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var summaries []RunSummary
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		s, err := summarise(path)
		if err != nil {
			continue
		}
		// Zero-entry files are aborted invocations (TUI opened and quit
		// at the picker); listing them buries real runs in `history`
		// and makes `cruft last` report a run where nothing happened.
		if s.Entries == 0 {
			continue
		}
		s.RunID = strings.TrimSuffix(e.Name(), ".jsonl")
		s.Path = path
		summaries = append(summaries, s)
	}
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].RunID > summaries[j].RunID })
	return summaries, nil
}

func summarise(path string) (RunSummary, error) {
	f, err := os.Open(path)
	if err != nil {
		return RunSummary{}, err
	}
	defer f.Close()
	out := RunSummary{PerCleaner: map[string]int64{}}
	dec := json.NewDecoder(f)
	// Compare as time.Time, not formatted strings: entries can carry
	// different UTC offsets (DST flip mid-run, travel), and lexical
	// comparison orders "01:00:00+05:30" after "22:00:00Z" despite it
	// being hours earlier.
	var first, last time.Time
	for dec.More() {
		var e cleaner.AuditEntry
		if err := dec.Decode(&e); err != nil {
			return out, err
		}
		out.Entries++
		if e.Success {
			out.Successes++
			out.BytesFreed += e.Bytes
			out.PerCleaner[e.Cleaner] += e.Bytes
		} else {
			out.Failures++
		}
		if first.IsZero() || e.Timestamp.Before(first) {
			first = e.Timestamp
		}
		if e.Timestamp.After(last) {
			last = e.Timestamp
		}
		if e.DryRun {
			out.DryRun = true
		}
	}
	if !first.IsZero() {
		out.FirstAction = first.UTC().Format(time.RFC3339)
		out.LastAction = last.UTC().Format(time.RFC3339)
	}
	return out, nil
}
