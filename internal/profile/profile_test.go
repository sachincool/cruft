package profile

import (
	"testing"

	"github.com/sachincool/cruft/internal/cleaner"
	_ "github.com/sachincool/cruft/internal/cleaner/all"
)

func TestParse(t *testing.T) {
	for _, tt := range []struct {
		in   string
		want Profile
	}{
		{"", Balanced},
		{"balanced", Balanced},
		{"conservative", Conservative},
		{"aggressive", Aggressive},
		{"unknown", Balanced},
	} {
		if got := Parse(tt.in); got != tt.want {
			t.Fatalf("Parse(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFilterSafetyProfiles(t *testing.T) {
	all := cleaner.All()
	balanced := Filter(all, Balanced)
	for _, c := range balanced {
		if c.Risky() {
			t.Fatalf("balanced included risky cleaner %q", c.Name())
		}
	}

	aggressive := Filter(all, Aggressive)
	if len(aggressive) != len(all) {
		t.Fatalf("aggressive selected %d cleaners, want %d", len(aggressive), len(all))
	}

	conservative := Filter(all, Conservative)
	if len(conservative) == 0 {
		t.Fatal("conservative selected no cleaners")
	}
	for _, c := range conservative {
		if c.Risky() {
			t.Fatalf("conservative included risky cleaner %q", c.Name())
		}
		if !fastRegenerate[c.Name()] {
			t.Fatalf("conservative included non-fast-regenerate cleaner %q", c.Name())
		}
	}
}
