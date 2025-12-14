package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Questline theme (CLI + TUI).
// Kept intentionally small: reusable styles and a few emojis.

const (
	IconQuest   = "🗺️"
	IconSparkle = "✨"
	IconPlus    = "➕"
	IconDone    = "✅"
	IconTrophy  = "🏆"
	IconBolt    = "⚡"
	IconInfo    = "ℹ️"
	IconWarn    = "⚠️"
	IconError   = "🧨"
	IconBox     = "📦"
	IconLoop    = "🔁"
	IconScroll  = "📜"
)

var (
	cPrimary = lipgloss.Color("63")  // blue
	cAccent  = lipgloss.Color("205") // magenta
	cGood    = lipgloss.Color("42")  // green
	cWarn    = lipgloss.Color("214") // orange
	cBad     = lipgloss.Color("196") // red
	cMuted   = lipgloss.Color("244") // gray
	cGold    = lipgloss.Color("220") // gold
)

var (
	Title = lipgloss.NewStyle().Bold(true).Foreground(cAccent)
	H2    = lipgloss.NewStyle().Bold(true).Foreground(cPrimary)
	Muted = lipgloss.NewStyle().Foreground(cMuted)
	Key   = lipgloss.NewStyle().Bold(true).Foreground(cPrimary)
	Good  = lipgloss.NewStyle().Bold(true).Foreground(cGood)
	Warn  = lipgloss.NewStyle().Bold(true).Foreground(cWarn)
	Bad   = lipgloss.NewStyle().Bold(true).Foreground(cBad)
	Gold  = lipgloss.NewStyle().Bold(true).Foreground(cGold)
	Dim   = lipgloss.NewStyle().Foreground(cMuted)

	Panel       = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(cMuted).Padding(0, 1)
	PanelTitle  = lipgloss.NewStyle().Bold(true).Foreground(cPrimary)
	SelectedRow = lipgloss.NewStyle().Bold(true).Foreground(cGold).Background(cPrimary)

	BadgeLevelUp = lipgloss.NewStyle().Bold(true).Foreground(cGold).Render("LEVEL UP")
)

func Heading(icon string, title string) string {
	icon = strings.TrimSpace(icon)
	if icon != "" {
		icon += " "
	}
	return Title.Render(icon + title)
}

func LabelValue(label string, value any) string {
	return fmt.Sprintf("%s %v", Key.Render(label+":"), value)
}

func StatusText(status string) string {
	s := strings.ToLower(strings.TrimSpace(status))
	switch s {
	case "done":
		return Good.Render("done")
	case "active":
		return H2.Render("active")
	case "pending":
		return Warn.Render("pending")
	default:
		return Muted.Render(status)
	}
}

func KindIcon(isProject bool, isHabit bool) string {
	if isProject {
		return IconBox
	}
	if isHabit {
		return IconLoop
	}
	return IconQuest
}
