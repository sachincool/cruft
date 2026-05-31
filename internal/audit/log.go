// Package audit writes a JSONL record per deletion attempt so users
// can reconstruct exactly what happened on any past run.
package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/sachincool/cruft/internal/cleaner"
)

// Log is a JSONL audit sink. A single goroutine serialises writes so
// concurrent cleaners can record entries safely.
type Log struct {
	dir      string
	runID    string
	mu       sync.Mutex
	f        *os.File
	enc      *json.Encoder
	disabled bool
}

// Open creates the audit directory if needed and opens a new file at
// <dir>/<runID>.jsonl.
func Open(dir, runID string) (*Log, error) {
	if dir == "" {
		return &Log{disabled: true}, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, runID+".jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	return &Log{
		dir:   dir,
		runID: runID,
		f:     f,
		enc:   json.NewEncoder(f),
	}, nil
}

// Record writes one entry. Safe for concurrent use.
func (l *Log) Record(e cleaner.AuditEntry) {
	if l == nil || l.disabled {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_ = l.enc.Encode(e)
}

// Close flushes and closes the file.
func (l *Log) Close() error {
	if l == nil || l.disabled {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.f != nil {
		err := l.f.Close()
		l.f = nil
		return err
	}
	return nil
}

// Path returns the audit log file path (empty if disabled).
func (l *Log) Path() string {
	if l == nil || l.disabled || l.f == nil {
		return ""
	}
	return l.f.Name()
}

// NopSink is an AuditSink that discards entries. Useful in tests and
// when audit is disabled.
type NopSink struct{}

func (NopSink) Record(cleaner.AuditEntry) {}
