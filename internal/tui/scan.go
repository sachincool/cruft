package tui

import (
	"fmt"
	"strings"
)

func renderScan(m *Model) string {
	var b strings.Builder
	b.WriteString(StyleTitle.Render("Scanning…"))
	b.WriteString("  ")
	b.WriteString(StyleMuted.Render(fmt.Sprintf("%d / %d cleaners", m.scanDone, m.scanTotal)))
	b.WriteString("\n\n")

	if m.lastScanned != "" {
		b.WriteString(StyleMuted.Render("checked  ") + StyleTitle.Render(m.lastScanned))
		b.WriteString("\n")
	}
	if m.foundBytes > 0 {
		b.WriteString(StyleMuted.Render("reclaimable so far  ") + StyleAccent.Render(HumanBytes(m.foundBytes)))
		b.WriteString("\n")
	}
	return b.String()
}
