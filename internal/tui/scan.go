package tui

import (
	"fmt"
	"strings"
)

func renderScan(m *Model) string {
	var b strings.Builder
	b.WriteString(StyleTitle.Render("Scanning…"))
	b.WriteString("\n")
	b.WriteString(StyleMuted.Render(fmt.Sprintf(
		"%d cleaners loaded. This shouldn't take long.", len(m.cleaners),
	)))
	b.WriteString("\n\n")
	b.WriteString(StyleMuted.Render("(no scan progress streaming yet — first pass returns when done)"))
	return b.String()
}
