package tui

import "strings"

func renderExecute(m *Model) string {
	var b strings.Builder
	b.WriteString(StyleTitle.Render("Cleaning…"))
	b.WriteString("\n")
	b.WriteString(StyleMuted.Render("(parallel execute; results come back when done)"))
	return b.String()
}
