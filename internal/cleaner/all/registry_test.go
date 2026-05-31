package all_test

import (
	"testing"

	"github.com/sachincool/cruft/internal/cleaner"
	_ "github.com/sachincool/cruft/internal/cleaner/all"
)

func TestRegisteredCleanersHaveTrustMetadata(t *testing.T) {
	cleaners := cleaner.All()
	if len(cleaners) < 20 {
		t.Fatalf("registered cleaners = %d, want at least 20", len(cleaners))
	}

	seen := map[string]bool{}
	for _, c := range cleaners {
		if c.Name() == "" {
			t.Fatal("cleaner with empty name")
		}
		if seen[c.Name()] {
			t.Fatalf("duplicate cleaner name %q", c.Name())
		}
		seen[c.Name()] = true

		if c.Description() == "" {
			t.Fatalf("%s: empty description", c.Name())
		}
		switch c.Category() {
		case cleaner.CategoryLangPkg, cleaner.CategoryIaC, cleaner.CategoryContainer, cleaner.CategorySystem:
		default:
			t.Fatalf("%s: unknown category %q", c.Name(), c.Category())
		}
		if c.Risky() && c.RiskReason() == "" {
			t.Fatalf("%s: risky cleaner must explain why", c.Name())
		}
		if !c.Risky() && c.RiskReason() != "" {
			t.Fatalf("%s: non-risky cleaner should not have risk reason %q", c.Name(), c.RiskReason())
		}
	}
}

func TestExpectedCleanersAreRegistered(t *testing.T) {
	want := []string{
		"npm", "pnpm", "yarn", "pip", "cargo", "gomod", "gem", "gradle", "maven",
		"terraform", "terragrunt", "pulumi",
		"docker", "docker-volumes", "colima",
		"homebrew", "xcode-derived", "xcode-archives", "xcode-simulators", "library-caches", "vscode", "jetbrains-caches", "jetbrains-system", "slack", "trash",
	}
	for _, name := range want {
		if cleaner.ByName(name) == nil {
			t.Fatalf("expected cleaner %q to be registered", name)
		}
	}
}
