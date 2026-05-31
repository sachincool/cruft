package tui

import "fmt"

// HumanBytes renders an int64 byte count as a short, scannable string.
func HumanBytes(n int64) string {
	if n < 0 {
		return "-" + HumanBytes(-n)
	}
	const (
		KB = 1 << 10
		MB = 1 << 20
		GB = 1 << 30
		TB = 1 << 40
	)
	switch {
	case n >= TB:
		return fmt.Sprintf("%.1f TB", float64(n)/float64(TB))
	case n >= GB:
		return fmt.Sprintf("%.1f GB", float64(n)/float64(GB))
	case n >= MB:
		return fmt.Sprintf("%.0f MB", float64(n)/float64(MB))
	case n >= KB:
		return fmt.Sprintf("%.0f KB", float64(n)/float64(KB))
	}
	return fmt.Sprintf("%d B", n)
}
