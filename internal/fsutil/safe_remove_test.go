package fsutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSafeRemoveRefusesOutsideWhitelist(t *testing.T) {
	root := t.TempDir()
	allowed := filepath.Join(root, "allowed")
	outside := filepath.Join(root, "outside")
	if err := os.MkdirAll(allowed, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outside, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SafeRemove(outside, []string{allowed}); !errors.Is(err, ErrOutsideWhitelist) {
		t.Fatalf("SafeRemove err = %v, want ErrOutsideWhitelist", err)
	}
	if _, err := os.Stat(outside); err != nil {
		t.Fatalf("outside file should remain: %v", err)
	}
}

func TestSafeRemoveRemovesInsideWhitelist(t *testing.T) {
	root := t.TempDir()
	allowed := filepath.Join(root, "allowed")
	target := filepath.Join(allowed, "cache")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := SafeRemove(target, []string{allowed}); err != nil {
		t.Fatalf("SafeRemove: %v", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("target should be removed, stat err=%v", err)
	}
}

func TestSafeRemoveDeletesSymlinkNotTarget(t *testing.T) {
	root := t.TempDir()
	allowed := filepath.Join(root, "allowed")
	target := filepath.Join(allowed, "real-data")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "f"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(allowed, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := SafeRemove(link, []string{allowed}); err != nil {
		t.Fatalf("SafeRemove: %v", err)
	}
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Fatalf("link should be removed, lstat err=%v", err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("symlink target must survive: %v", err)
	}
}

func TestSafeRemoveDeletesDanglingSymlink(t *testing.T) {
	root := t.TempDir()
	allowed := filepath.Join(root, "allowed")
	if err := os.MkdirAll(allowed, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(allowed, "dangling")
	if err := os.Symlink(filepath.Join(root, "no-such-target"), link); err != nil {
		t.Fatal(err)
	}
	if err := SafeRemove(link, []string{allowed}); err != nil {
		t.Fatalf("SafeRemove: %v", err)
	}
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Fatalf("dangling link should be removed, lstat err=%v", err)
	}
}

func TestIsUnderResolvesSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	allowed := filepath.Join(root, "allowed")
	outside := filepath.Join(root, "outside")
	if err := os.MkdirAll(allowed, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(allowed, "link")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	ok, err := IsUnder(link, []string{allowed})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("symlink escape should not be treated as under whitelist")
	}
}
