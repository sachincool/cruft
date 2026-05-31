package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sachincool/cruft/internal/cleaner"
	"github.com/sachincool/cruft/internal/fsutil"
	"github.com/sachincool/cruft/internal/runner"
)

type phase int

const (
	phaseScan phase = iota
	phaseSelect
	phaseConfirm
	phaseExecute
	phaseSummary
)

// Model is the bubbletea root model.
type Model struct {
	ctx       context.Context
	runner    *runner.Runner
	cleaners  []cleaner.Cleaner
	phase     phase
	scanRes   []runner.ScanResult
	execRes   []runner.ExecResult
	cursor    int
	beforeFS  int64
	afterFS   int64
	width     int
	height    int
	scanDone  int
	scanTotal int
	// Live-scan plumbing: cleaners stream their results back over scanCh
	// as they finish. scanIdx maps a cleaner name to its stable slot in
	// scanRes so the select view keeps a deterministic order.
	scanCh      chan runner.ScanResult
	scanIdx     map[string]int
	lastScanned string
	foundBytes  int64
	err         error
	// helpFor is the cleaner whose Description is currently shown in
	// the help overlay (toggled by `?`). nil = panel hidden.
	helpFor cleaner.Cleaner
}

// NewModel returns the root TUI model. Pass the prepared runner and
// the slice of cleaners to scan.
func NewModel(ctx context.Context, r *runner.Runner, cs []cleaner.Cleaner) *Model {
	return &Model{
		ctx:       ctx,
		runner:    r,
		cleaners:  cs,
		phase:     phaseScan,
		scanTotal: len(cs),
		beforeFS:  fsutil.FreeBytes("/"),
	}
}

func (m *Model) Init() tea.Cmd { return m.startScan() }

func (m *Model) startScan() tea.Cmd {
	// Pre-fill a slot per cleaner so results land in a stable order
	// regardless of which scanner finishes first.
	m.scanRes = make([]runner.ScanResult, len(m.cleaners))
	m.scanIdx = make(map[string]int, len(m.cleaners))
	for i, c := range m.cleaners {
		m.scanRes[i].Cleaner = c
		m.scanIdx[c.Name()] = i
	}
	m.scanCh = make(chan runner.ScanResult, len(m.cleaners))
	ch := m.scanCh
	go func() {
		// Each cleaner calls progress(res) exactly once as it finishes.
		_, _ = m.runner.Scan(m.ctx, m.cleaners, func(r runner.ScanResult) { ch <- r })
		close(ch)
	}()
	return waitForScan(ch)
}

// waitForScan blocks on the next streamed result, turning it into a
// message. A closed channel means the scan is done.
func waitForScan(ch chan runner.ScanResult) tea.Cmd {
	return func() tea.Msg {
		r, ok := <-ch
		if !ok {
			return scanCompleteMsg{}
		}
		return scanProgressMsg{Result: r}
	}
}

func (m *Model) startExecute() tea.Cmd {
	return func() tea.Msg {
		results := m.runner.Execute(m.ctx, m.scanRes, nil)
		return execCompleteMsg{Results: results}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	case scanProgressMsg:
		if i, ok := m.scanIdx[msg.Result.Cleaner.Name()]; ok {
			m.scanRes[i] = msg.Result
		}
		m.scanDone++
		m.lastScanned = msg.Result.Cleaner.Name()
		m.foundBytes += msg.Result.TotalBytes
		return m, waitForScan(m.scanCh)
	case scanCompleteMsg:
		m.scanDone = m.scanTotal
		m.phase = phaseSelect
		return m, nil
	case execCompleteMsg:
		m.execRes = msg.Results
		m.afterFS = fsutil.FreeBytes("/")
		m.phase = phaseSummary
		return m, nil
	case errMsg:
		m.err = msg.Err
		return m, nil
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	}
	switch m.phase {
	case phaseSelect:
		return m.handleSelectKey(msg)
	case phaseConfirm:
		switch msg.String() {
		case "y", "enter":
			m.phase = phaseExecute
			return m, m.startExecute()
		case "n", "esc":
			m.phase = phaseSelect
			return m, nil
		}
	case phaseSummary:
		switch msg.String() {
		case "enter", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *Model) handleSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visible := m.visibleScans()
	if len(visible) == 0 {
		if msg.String() == "enter" {
			return m, tea.Quit
		}
		return m, nil
	}
	// If the help panel is open, any key closes it. Don't let nav
	// keys leak through and move the cursor under a hidden panel.
	if m.helpFor != nil {
		m.helpFor = nil
		return m, nil
	}
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(visible)-1 {
			m.cursor++
		}
	case "?":
		// Open help panel for the focused cleaner. We surface the
		// cleaner's plain-language Description() so users learn what
		// "risky" / what each cleaner actually does without quitting.
		idx := visible[m.cursor]
		m.helpFor = m.scanRes[idx].Cleaner
	case "d":
		// Toggle dry-run mid-session. Lets a user start with a real
		// confirm in mind, then drop to preview-only at the last second
		// (or vice versa) without quitting and re-launching.
		m.runner.SetDryRun(!m.runner.IsDryRun())
	case "s":
		// Toggle safe-mode (tombstone) mid-session. Surfaces a feature
		// users couldn't otherwise discover from inside the TUI.
		m.runner.SetSafe(!m.runner.UsesTombstone())
	case " ":
		// Toggle all approved flags for the cleaner at cursor.
		idx := visible[m.cursor]
		toggleApprovedForScan(&m.scanRes[idx])
	case "a":
		// Toggle approvedness across the whole list.
		anyApproved := false
		for _, i := range visible {
			for _, f := range m.scanRes[i].Findings {
				if f.Approved {
					anyApproved = true
					break
				}
			}
		}
		for _, i := range visible {
			for j := range m.scanRes[i].Findings {
				m.scanRes[i].Findings[j].Approved = !anyApproved
			}
		}
	case "enter":
		m.phase = phaseConfirm
	}
	return m, nil
}

func toggleApprovedForScan(s *runner.ScanResult) {
	anyApproved := false
	for _, f := range s.Findings {
		if f.Approved {
			anyApproved = true
			break
		}
	}
	for i := range s.Findings {
		s.Findings[i].Approved = !anyApproved
	}
}

// visibleScans returns indices of scans that have findings to show.
func (m *Model) visibleScans() []int {
	var out []int
	for i, s := range m.scanRes {
		if s.NotInstalled || s.BusyProcess != "" {
			continue
		}
		if len(s.Findings) == 0 {
			continue
		}
		out = append(out, i)
	}
	return out
}

func (m *Model) selectedBytes() int64 {
	var n int64
	for _, s := range m.scanRes {
		for _, f := range s.Findings {
			if f.Approved {
				n += f.Bytes
			}
		}
	}
	return n
}

func (m *Model) View() string {
	var b strings.Builder
	b.WriteString("  ")
	b.WriteString(StyleAccent.Render(BannerMark))
	b.WriteString(StyleDim.Render(BannerFade))
	b.WriteString("\n  ")
	b.WriteString(StyleMuted.Render(BannerTagline))
	b.WriteString("\n\n")

	switch m.phase {
	case phaseScan:
		b.WriteString(renderScan(m))
	case phaseSelect:
		b.WriteString(renderSelect(m))
	case phaseConfirm:
		b.WriteString(renderConfirm(m))
	case phaseExecute:
		b.WriteString(renderExecute(m))
	case phaseSummary:
		b.WriteString(renderSummary(m))
	}

	// Help overlay (? key) — sits between the list and the footer so
	// the user sees the cleaner's plain-language description without
	// quitting the TUI. Any key dismisses.
	if m.helpFor != nil {
		b.WriteString("\n\n")
		b.WriteString(renderHelpPanel(m.helpFor))
	}

	b.WriteString("\n")
	b.WriteString(renderFooter(m))
	return b.String()
}

// renderHelpPanel formats the focused cleaner's docs into a bordered
// block. Lead with risk if risky — that's the question users have.
func renderHelpPanel(c cleaner.Cleaner) string {
	var inner strings.Builder
	inner.WriteString(StyleTitle.Render(c.Name()))
	inner.WriteString(StyleMuted.Render("  · " + string(c.Category())))
	inner.WriteString("\n")
	if c.Risky() {
		reason := c.RiskReason()
		if reason == "" {
			reason = "unchecked by default"
		}
		inner.WriteString(StyleWarn.Render("⚠  risky · " + reason))
		inner.WriteString("\n")
	}
	inner.WriteString("\n")
	inner.WriteString(c.Description())
	inner.WriteString("\n\n")
	inner.WriteString(StyleMuted.Render("press any key to close"))
	return StyleBanner.Render(inner.String())
}

// renderFooter shows two distinct, labelled values:
//
//	Will free  = sum of ticked rows (exactly what you see above)
//	Disk free  = current state → resulting state (secondary context)
//
// Earlier versions led with "Predicted: <future free disk>" which confused
// users on first run — the rows summed to ~10 GB but the headline said
// 30 GB, leaving them unable to reconcile the math. The actionable
// number (Will free) is now the headline; the disk-free delta is context.
func renderFooter(m *Model) string {
	selected := m.selectedBytes()
	predicted := m.beforeFS + selected
	var modeNote string
	if m.runner != nil && m.runner.IsDryRun() {
		modeNote = StyleAccent.Render("DRY RUN") + " — nothing will be deleted"
	} else {
		modeNote = StyleDanger.Render("LIVE") + " — files will be tombstoned/deleted"
	}
	line := fmt.Sprintf(
		"Will free: %s   ·   Disk free: %s → %s   ·   %s",
		StyleAccent.Render(HumanBytes(selected)),
		StyleTitle.Render(HumanBytes(m.beforeFS)),
		StyleAccent.Render(HumanBytes(predicted)),
		modeNote,
	)
	return StyleFooter.Render(line)
}
