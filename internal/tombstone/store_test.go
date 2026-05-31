package tombstone

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuryAndRestoreFile(t *testing.T) {
	root := t.TempDir()
	srcDir := filepath.Join(root, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(srcDir, "cache.txt")
	if err := os.WriteFile(src, []byte("cache"), 0o644); err != nil {
		t.Fatal(err)
	}

	store := New(filepath.Join(root, "tombstone"))
	dst, err := store.Bury("run-1", "test", src)
	if err != nil {
		t.Fatalf("Bury: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("source should be moved, stat err=%v", err)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("tombstone missing: %v", err)
	}

	restored, skipped, err := store.Restore("run-1")
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if len(skipped) != 0 || len(restored) != 1 || restored[0] != src {
		t.Fatalf("restored=%v skipped=%v", restored, skipped)
	}
	got, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "cache" {
		t.Fatalf("restored content = %q", got)
	}
}

func TestRestoreSkipsExistingPath(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "cache.txt")
	if err := os.WriteFile(src, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := New(filepath.Join(root, "tombstone"))
	if _, err := store.Bury("run-1", "test", src); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, skipped, err := store.Restore("run-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) != 1 || skipped[0] != src {
		t.Fatalf("skipped=%v", skipped)
	}
}
