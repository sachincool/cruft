package fsutil

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
)

// HasLocalSnapshots reports whether the volume at mount has APFS / Time
// Machine local snapshots. These freeze the volume's state hourly, so a
// file you delete stays referenced by the snapshot — its blocks become
// "purgeable" and don't return to free space until the snapshot is
// thinned. That's why a successful delete can leave Storage unchanged.
//
// macOS-only signal: returns false anywhere tmutil isn't on $PATH.
func HasLocalSnapshots(ctx context.Context, mount string) bool {
	if !CommandExists("tmutil") {
		return false
	}
	out, err := exec.CommandContext(ctx, "tmutil", "listlocalsnapshots", mount).Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "com.apple.TimeMachine")
}

// thinLocalSnapshots asks macOS to thin local snapshots on mount until up
// to purgeBytes has been reclaimed, at the highest urgency (4). No-op
// (nil) when tmutil is absent or the target is non-positive. Does not
// require root.
func thinLocalSnapshots(ctx context.Context, mount string, purgeBytes int64) error {
	if !CommandExists("tmutil") || purgeBytes <= 0 {
		return nil
	}
	return exec.CommandContext(ctx, "tmutil", "thinlocalsnapshots", mount,
		strconv.FormatInt(purgeBytes, 10), "4").Run()
}

// ReclaimLocalSnapshots returns snapshot-pinned space to the volume at
// mount and reports how many bytes of free space that recovered.
//
// After a cleanup, blocks freed by deletion can stay "purgeable" because
// local snapshots still reference them — pure waste once the live files
// are gone. The thin target is bounded by targetBytes (what this run
// deleted): tmutil drops the oldest snapshots until that much space is
// released, keeping newer restore points intact. An unbounded thin would
// destroy every local restore point on the machine, far beyond what the
// user approved. Real Time Machine backups on a destination drive are
// never touched.
//
// No-op (0) when tmutil is absent, there are no local snapshots, or
// targetBytes is non-positive (nothing was deleted, nothing to unpin).
func ReclaimLocalSnapshots(ctx context.Context, mount string, targetBytes int64) int64 {
	if targetBytes <= 0 || !HasLocalSnapshots(ctx, mount) {
		return 0
	}
	before := FreeBytes(mount)
	_ = thinLocalSnapshots(ctx, mount, targetBytes)
	return max(FreeBytes(mount)-before, 0)
}
