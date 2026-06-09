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
// are gone. We thin aggressively (highest urgency, target larger than any
// real disk) so a single run actually hands the space back to the OS,
// including space pinned by deletions from earlier runs. tmutil keeps the
// newest restore points it can while meeting the target, and real Time
// Machine backups on a destination drive are never touched.
//
// No-op (0) when tmutil is absent or there are no local snapshots.
func ReclaimLocalSnapshots(ctx context.Context, mount string) int64 {
	if !HasLocalSnapshots(ctx, mount) {
		return 0
	}
	before := FreeBytes(mount)
	// 1<<50 (~1 PiB) is far above any real volume, so tmutil reclaims
	// every thinnable snapshot rather than stopping at a byte target.
	_ = thinLocalSnapshots(ctx, mount, 1<<50)
	return max(FreeBytes(mount)-before, 0)
}
