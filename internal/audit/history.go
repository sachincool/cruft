package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
		ts := e.Timestamp.Format("2006-01-02T15:04:05Z07:00")
		if out.FirstAction == "" || ts < out.FirstAction {
			out.FirstAction = ts
		}
		if ts > out.LastAction {
			out.LastAction = ts
		}
		if e.DryRun {
			out.DryRun = true
		}
	}
	return out, nil
}
