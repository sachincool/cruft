package cleaner

import (
	"fmt"
	"sort"
	"sync"
)

var (
	regMu sync.RWMutex
	reg   = map[string]Cleaner{}
)

// Register adds a cleaner to the global registry. Called from init()
// in each concrete cleaner file. Panics on duplicate Name() — a
// programmer error worth failing loud.
func Register(c Cleaner) {
	regMu.Lock()
	defer regMu.Unlock()
	if _, exists := reg[c.Name()]; exists {
		panic(fmt.Sprintf("cleaner: duplicate registration for %q", c.Name()))
	}
	reg[c.Name()] = c
}

// All returns every registered cleaner, sorted by category then name.
func All() []Cleaner {
	regMu.RLock()
	defer regMu.RUnlock()
	out := make([]Cleaner, 0, len(reg))
	for _, c := range reg {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Category() != out[j].Category() {
			return out[i].Category() < out[j].Category()
		}
		return out[i].Name() < out[j].Name()
	})
	return out
}

// ByName looks up a single cleaner. Returns nil if not found.
func ByName(name string) Cleaner {
	regMu.RLock()
	defer regMu.RUnlock()
	return reg[name]
}

// ByNames resolves a list of names. Unknown names are returned in the
// second slice so callers can report them.
func ByNames(names []string) (found []Cleaner, unknown []string) {
	for _, n := range names {
		if c := ByName(n); c != nil {
			found = append(found, c)
		} else {
			unknown = append(unknown, n)
		}
	}
	return
}
