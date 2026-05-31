package tui

import (
	"fmt"
	"strings"

	"github.com/sachincool/cruft/internal/cleaner"
	"github.com/sachincool/cruft/internal/runner"
)

func renderSelect(m *Model) string {
	var b strings.Builder
	b.WriteString(StyleTitle.Render("What to clean"))
	b.WriteString("\n")
	b.WriteString(StyleMuted.Render(
		"space toggle · a toggle-all · ? what is this · d preview · s safe-mode · enter confirm · q quit",
	))
	// Status pills so the user sees which modes are on right now and
	// what `d`/`s` would flip them to. Without this, the toggle hints
	// feel abstract — you can't tell if you're already in preview.
	dryPill := StyleMuted.Render("[d] preview off")
	if m.runner.IsDryRun() {
		dryPill = StyleAccent.Render("[d] preview ON")
	}
	safePill := StyleMuted.Render("[s] safe-mode off (delete is final)")
	if m.runner.UsesTombstone() {
		safePill = StyleAccent.Render("[s] safe-mode ON (7d recycle)")
	}
	b.WriteString("\n")
	b.WriteString("  " + dryPill + "   " + safePill)
	b.WriteString("\n")

	// Group visible scans by category.
	groups := map[cleaner.Category][]int{}
	order := []cleaner.Category{
		cleaner.CategoryLangPkg, cleaner.CategoryIaC,
		cleaner.CategoryContainer, cleaner.CategorySystem,
	}
	visible := m.visibleScans()
	for _, idx := range visible {
		c := m.scanRes[idx].Cleaner.Category()
		groups[c] = append(groups[c], idx)
	}

	// Render each group in order.
	for _, cat := range order {
		idxs := groups[cat]
		if len(idxs) == 0 {
			continue
		}
		b.WriteString(StyleCategory.Render(strings.ToUpper(cat.Title())))
		b.WriteString("\n")
		for _, i := range idxs {
			s := m.scanRes[i]
			cursor := "  "
			rowIdx := indexOf(visible, i)
			if rowIdx == m.cursor {
				cursor = "▶ "
			}
			checked := allApproved(s)
			box := "[ ]"
			if checked {
				box = "[" + StyleAccent.Render("✓") + "]"
			}
			risky := ""
			if s.Cleaner.Risky() {
				risky = " " + StyleWarn.Render("!")
			}
			line := fmt.Sprintf(
				"%s%s %-20s %s%s   %s",
				cursor,
				box,
				s.Cleaner.Name(),
				StyleAccent.Render(HumanBytes(s.TotalBytes)),
				risky,
				StyleMuted.Render(reasonHint(s)),
			)
			b.WriteString(line + "\n")
		}
	}

	// Show ignored cleaners as a fold-line.
	var skippedBusy, notInstalled, emptyScans int
	for _, s := range m.scanRes {
		switch {
		case s.NotInstalled:
			notInstalled++
		case s.BusyProcess != "":
			skippedBusy++
		case len(s.Findings) == 0 && s.Err == nil:
			emptyScans++
		}
	}
	b.WriteString("\n")
	b.WriteString(StyleMuted.Render(fmt.Sprintf(
		"·  %d not installed   ·  %d already clean   ·  %d skipped (in use)",
		notInstalled, emptyScans, skippedBusy,
	)))
	return b.String()
}

func allApproved(s runner.ScanResult) bool {
	if len(s.Findings) == 0 {
		return false
	}
	for _, f := range s.Findings {
		if !f.Approved {
			return false
		}
	}
	return true
}

func indexOf(xs []int, target int) int {
	for i, x := range xs {
		if x == target {
			return i
		}
	}
	return -1
}

// reasonHint picks the most informative one-line note for a scan row.
func reasonHint(s runner.ScanResult) string {
	if len(s.Findings) == 0 {
		return ""
	}
	if len(s.Findings) == 1 {
		return s.Findings[0].Reason
	}
	return fmt.Sprintf("%d paths", len(s.Findings))
}
