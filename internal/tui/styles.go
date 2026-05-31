// Package tui implements the bubbletea TUI for cruft.
//
// Brand palette: stark monochrome + electric green accent for
// "freed/reclaimed" numbers. Brutalist wordmark, manifesto vibe.
package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Brand colours.
	colorAccent = lipgloss.Color("#00FF66") // electric green — "freed"
	colorWarn   = lipgloss.Color("#FFD400") // mustard — risky / heads-up
	colorDanger = lipgloss.Color("#FF3B30") // red — destructive mode active
	colorMuted  = lipgloss.Color("#6C6C6C")
	colorDim    = lipgloss.Color("#444444")
	colorFG     = lipgloss.Color("#FFFFFF")

	StyleTitle  = lipgloss.NewStyle().Foreground(colorFG).Bold(true)
	StyleAccent = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	StyleWarn   = lipgloss.NewStyle().Foreground(colorWarn).Bold(true)
	StyleDanger = lipgloss.NewStyle().Foreground(colorDanger).Bold(true)
	StyleMuted  = lipgloss.NewStyle().Foreground(colorMuted)
	StyleDim    = lipgloss.NewStyle().Foreground(colorDim)

	StyleCategory = lipgloss.NewStyle().
			Foreground(colorFG).
			Bold(true).
			MarginTop(1)

	StyleCheckbox = lipgloss.NewStyle().Foreground(colorFG)
	StyleRisky    = lipgloss.NewStyle().Foreground(colorWarn)

	StyleBanner = lipgloss.NewStyle().
			Foreground(colorFG).
			Bold(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorAccent).
			Padding(0, 1)

	StyleDangerBanner = lipgloss.NewStyle().
				Foreground(colorDanger).
				Bold(true).
				BorderStyle(lipgloss.DoubleBorder()).
				BorderForeground(colorDanger).
				Padding(0, 1)

	StyleFooter = lipgloss.NewStyle().
			Foreground(colorMuted).
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorDim).
			PaddingTop(1).
			MarginTop(1)
)

// Banner is the wordmark shown above the TUI.
//
// Modern terminal-CLI aesthetic: letterspaced caps in the brand colour,
// trailing fade ramp as the "eaten-away" metaphor. Two lines, not five
// — keeps vertical room for actual content.
const (
	BannerMark    = "C R U F T"
	BannerFade    = " ░▒▓"
	BannerTagline = "decruft your laptop"
)

// Mark is a single-character brand glyph for tight contexts
// (compact summaries, error prefixes).
const Mark = "▎"
