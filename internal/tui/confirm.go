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
	// Gate on UsesTombstone, not TombstoneRoot: the root stays set after
	// safe-mode is toggled off, and promising `cruft restore` next to a
	// "PERMANENTLY DELETED" banner would be a lie.
	if m.runner.UsesTombstone() && !dry {
		b.WriteString(StyleMuted.Render("Tombstone:  "+m.runner.TombstoneRoot()+"  (restore with `cruft restore`)") + "\n")
	}
	if m.opts.ReclaimSnapshots && !dry {
		b.WriteString(StyleMuted.Render("Time Machine local snapshots will be thinned to release the freed space.") + "\n")
	}
	b.WriteString("\n")
	b.WriteString(StyleTitle.Render("[y] confirm   [n] back"))
	return strings.TrimRight(b.String(), "\n")
}
