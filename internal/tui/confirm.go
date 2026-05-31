package tui

import (
	"fmt"
	"strings"
)

func renderConfirm(m *Model) string {
	var b strings.Builder
	b.WriteString(StyleTitle.Render("Ready?"))
	b.WriteString("\n\n")
	selected := m.selectedBytes()
	dry := m.runner.IsDryRun()
	if dry {
		b.WriteString(StyleBanner.Render(fmt.Sprintf(
			"DRY RUN — nothing will be deleted. Would free %s.",
			HumanBytes(selected),
		)))
	} else if m.runner.UsesTombstone() {
		b.WriteString(StyleBanner.Render(fmt.Sprintf(
			"LIVE — files will be moved to tombstone (recoverable). About to reclaim %s.",
			HumanBytes(selected),
		)))
	} else {
		b.WriteString(StyleDangerBanner.Render(fmt.Sprintf(
			"LIVE — files will be PERMANENTLY DELETED. About to reclaim %s.",
			HumanBytes(selected),
		)))
	}
	b.WriteString("\n\n")
	if path := m.runner.AuditPath(); path != "" {
		b.WriteString(StyleMuted.Render("Audit log: "+path) + "\n")
	}
	if root := m.runner.TombstoneRoot(); root != "" && !dry {
		b.WriteString(StyleMuted.Render("Tombstone:  "+root+"  (restore with `cruft restore`)") + "\n")
	}
	b.WriteString("\n")
	b.WriteString(StyleTitle.Render("[y] confirm   [n] back"))
	return strings.TrimRight(b.String(), "\n")
}
