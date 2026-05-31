package runner

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
)

// Budget tracks bytes reclaimed against an optional cap. Calls to
// Add return true while there's room left; false once the cap is met.
// A zero Cap means unlimited.
type Budget struct {
	cap   int64
	freed atomic.Int64
}

func NewBudget(cap int64) *Budget { return &Budget{cap: cap} }

// Add records bytes freed; returns false if the cap has been reached.
func (b *Budget) Add(bytes int64) bool {
	if b == nil || b.cap <= 0 {
		if b != nil {
			b.freed.Add(bytes)
		}
		return true
	}
	b.freed.Add(bytes)
	return b.freed.Load() < b.cap
}

// Exhausted returns true if the cap has been reached.
func (b *Budget) Exhausted() bool {
	if b == nil || b.cap <= 0 {
		return false
	}
	return b.freed.Load() >= b.cap
}

// ParseSize parses strings like "5GB", "500MB", "1.5G", "200000" (raw bytes).
func ParseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	upper := strings.ToUpper(s)
	multipliers := []struct {
		suffix string
		mul    float64
	}{
		{"TB", 1 << 40}, {"GB", 1 << 30}, {"MB", 1 << 20}, {"KB", 1 << 10},
		{"T", 1 << 40}, {"G", 1 << 30}, {"M", 1 << 20}, {"K", 1 << 10},
		{"B", 1},
	}
	for _, m := range multipliers {
		if strings.HasSuffix(upper, m.suffix) {
			num := strings.TrimSuffix(upper, m.suffix)
			f, err := strconv.ParseFloat(strings.TrimSpace(num), 64)
			if err != nil {
				return 0, fmt.Errorf("invalid size %q: %w", s, err)
			}
			return int64(f * m.mul), nil
		}
	}
	// Bare number → bytes.
	f, err := strconv.ParseFloat(upper, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q: %w", s, err)
	}
	return int64(f), nil
}
