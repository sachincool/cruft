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

// ThinLocalSnapshots asks macOS to thin local snapshots on mount until up
// to purgeBytes has been reclaimed, at the highest urgency (4). tmutil
// thins only as many snapshots as needed to hit the target, so passing
// the just-deleted byte count recovers that space while keeping the
// newest restore points it can. No-op (nil) when tmutil is absent or the
// target is non-positive. Does not require root.
func ThinLocalSnapshots(ctx context.Context, mount string, purgeBytes int64) error {
	if !CommandExists("tmutil") || purgeBytes <= 0 {
		return nil
	}
	return exec.CommandContext(ctx, "tmutil", "thinlocalsnapshots", mount,
		strconv.FormatInt(purgeBytes, 10), "4").Run()
}
