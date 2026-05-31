package tui

import (
	"fmt"
	"strings"
)

func renderSummary(m *Model) string {
	var b strings.Builder
	b.WriteString(StyleTitle.Render("Done."))
	b.WriteString("\n\n")

	var totalFreed int64
	for _, r := range m.execRes {
		totalFreed += r.Result.BytesFreed
	}
	dry := runnerIsDryRun(m.runner)

	if dry {
		b.WriteString(fmt.Sprintf(
			"Would have freed %s. (dry run)\n\n",
			StyleAccent.Render(HumanBytes(totalFreed)),
		))
	} else {
		b.WriteString(fmt.Sprintf(
			"Freed %s.\n\n",
			StyleAccent.Render(HumanBytes(totalFreed)),
		))
	}

	// Per-cleaner table.
	for _, r := range m.execRes {
		if r.Result.Findings == 0 {
			continue
		}
		status := StyleAccent.Render("✓")
		if r.Result.Failed > 0 {
			status = StyleDanger.Render("✗")
		}
		b.WriteString(fmt.Sprintf(
			"  %s %-22s %s\n",
			status,
			r.Cleaner.Name(),
			StyleAccent.Render(HumanBytes(r.Result.BytesFreed)),
		))
	}

	// Disk delta.
	//
	// In dry-run, nothing was actually deleted, so the *measured*
	// before/after (m.afterFS - m.beforeFS) is ~0 plus noise from
	// other processes. Showing that contradicted the headline ("Would
	// have freed 10.7 GB") and the footer. We now mirror the headline
	// math in dry-run by showing the predicted post-state.
	afterDisplay := m.afterFS
	deltaDisplay := m.afterFS - m.beforeFS
	deltaWord := "freed"
	if dry {
		afterDisplay = m.beforeFS + totalFreed
		deltaDisplay = totalFreed
		deltaWord = "would be freed"
	}
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(
		"Disk free:  %s → %s   (%s %s)\n",
		HumanBytes(m.beforeFS),
		StyleAccent.Render(HumanBytes(afterDisplay)),
		StyleAccent.Render(HumanBytes(deltaDisplay)),
		deltaWord,
	))

	if path := m.runner.AuditPath(); path != "" {
		b.WriteString("\n")
		b.WriteString(StyleMuted.Render("Receipt: " + path))
		b.WriteString("\n")
	}

	// Next-step bridge from TUI → CLI. Each line is a command the
	// user can paste into a fresh terminal to continue the workflow.
	// Without this, the TUI ends in a vacuum and users don't discover
	// `last`, `history`, or `restore` until they search docs.
	b.WriteString("\n")
	b.WriteString(StyleTitle.Render("What's next"))
	b.WriteString("\n")
	if m.runner.UsesTombstone() && !dry {
		b.WriteString(fmt.Sprintf(
			"  %s   undo this run (within 7 days)\n",
			StyleAccent.Render("cruft restore "+m.runner.RunID()),
		))
	}
	b.WriteString("  " + StyleAccent.Render("cruft last") +
		StyleMuted.Render("                          per-cleaner detail for this run") + "\n")
	b.WriteString("  " + StyleAccent.Render("cruft history") +
		StyleMuted.Render("                       all past runs") + "\n")
	b.WriteString("  " + StyleAccent.Render("cruft") +
		StyleMuted.Render("                               run again") + "\n")
	if !m.runner.UsesTombstone() && !dry {
		b.WriteString("  " + StyleMuted.Render("(tip: re-run with ") +
			StyleAccent.Render("--safe") +
			StyleMuted.Render(" if you want a 7-day undo window)") + "\n")
	}

	b.WriteString("\n")
	b.WriteString(StyleMuted.Render("Press enter to quit."))
	return strings.TrimRight(b.String(), "\n")
}
