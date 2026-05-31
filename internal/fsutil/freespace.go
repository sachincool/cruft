package fsutil

import "syscall"

// FreeBytes returns the free space (in bytes) on the filesystem
// containing path. On error returns 0.
func FreeBytes(path string) int64 {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0
	}
	// Bavail = blocks available to unprivileged users.
	return int64(st.Bavail) * int64(st.Bsize)
}
