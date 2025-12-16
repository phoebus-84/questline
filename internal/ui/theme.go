package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Questline theme (CLI + TUI).
// Retro gaming + hacker aesthetic with phosphor green/amber colors.

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

// Retro terminal colors (phosphor green/amber CRT aesthetic)
var (
	cPhosphor  = lipgloss.Color("46")  // bright phosphor green
	cMatrix    = lipgloss.Color("34")  // darker matrix green
	cAmber     = lipgloss.Color("214") // amber/gold
	cCyan      = lipgloss.Color("51")  // bright cyan
	cMagenta   = lipgloss.Color("201") // bright magenta
	cRed       = lipgloss.Color("196") // red for warnings
	cDim       = lipgloss.Color("239") // very dim gray
	cDark      = lipgloss.Color("232") // nearly black
	cTermGreen = lipgloss.Color("40")  // terminal green

	// Legacy compatibility aliases
	cPrimary = cPhosphor
	cAccent  = cAmber
	cGood    = cPhosphor
	cWarn    = cAmber
	cBad     = cRed
	cMuted   = lipgloss.Color("244")
	cGold    = cAmber
)

var (
	Title = lipgloss.NewStyle().Bold(true).Foreground(cAmber)
	H2    = lipgloss.NewStyle().Bold(true).Foreground(cPhosphor)
	Muted = lipgloss.NewStyle().Foreground(cMuted)
	Key   = lipgloss.NewStyle().Bold(true).Foreground(cPhosphor)
	Good  = lipgloss.NewStyle().Bold(true).Foreground(cPhosphor)
	Warn  = lipgloss.NewStyle().Bold(true).Foreground(cAmber)
	Bad   = lipgloss.NewStyle().Bold(true).Foreground(cRed)
	Gold  = lipgloss.NewStyle().Bold(true).Foreground(cAmber)
	Dim   = lipgloss.NewStyle().Foreground(cDim)

	// Retro terminal styles
	Terminal     = lipgloss.NewStyle().Foreground(cPhosphor)
	TerminalDim  = lipgloss.NewStyle().Foreground(cMatrix)
	TerminalBold = lipgloss.NewStyle().Bold(true).Foreground(cPhosphor)
	Cyber        = lipgloss.NewStyle().Bold(true).Foreground(cCyan)
	Neon         = lipgloss.NewStyle().Bold(true).Foreground(cMagenta)

	// Panel with retro border
	Panel = lipgloss.NewStyle().
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(cMatrix).
		Padding(0, 1)

	PanelTitle  = lipgloss.NewStyle().Bold(true).Foreground(cAmber)
	SelectedRow = lipgloss.NewStyle().Bold(true).Foreground(cDark).Background(cPhosphor)

	BadgeLevelUp = lipgloss.NewStyle().Bold(true).Foreground(cAmber).Render("▲▲▲ LEVEL UP ▲▲▲")

	// Scanline effect (simulated with dim lines)
	Scanline = lipgloss.NewStyle().Foreground(cDim)
)

// ASCII art header for retro feel
var ASCIIHeader = `
╔═══════════════════════════════════════════════════════════════════════════╗
║  ██████  ██    ██ ███████ ███████ ████████ ██      ██ ███    ██ ███████   ║
║ ██    ██ ██    ██ ██      ██         ██    ██      ██ ████   ██ ██        ║
║ ██    ██ ██    ██ █████   ███████    ██    ██      ██ ██ ██  ██ █████     ║
║ ██ ▄▄ ██ ██    ██ ██           ██    ██    ██      ██ ██  ██ ██ ██        ║
║  ██████   ██████  ███████ ███████    ██    ███████ ██ ██   ████ ███████   ║
║      ▀▀                                                                   ║
╚═══════════════════════════════════════════════════════════════════════════╝`

// Compact ASCII header
var ASCIIHeaderCompact = `▓▓▓ QUESTLINE ▓▓▓`

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

// RetroBox wraps content in a retro terminal box
func RetroBox(title string, content string, width int) string {
	if width < 10 {
		width = 40
	}
	titleLen := lipgloss.Width(title)
	innerW := width - 4 // Account for borders and padding

	// Top border with title
	topLeft := "╔═"
	topRight := "═╗"
	titlePart := "[ " + title + " ]"
	remainingTop := innerW - titleLen - 4
	if remainingTop < 0 {
		remainingTop = 0
	}
	top := topLeft + strings.Repeat("═", remainingTop/2) + titlePart + strings.Repeat("═", (remainingTop+1)/2) + topRight

	// Bottom border
	bottom := "╚" + strings.Repeat("═", width-2) + "╝"

	// Wrap content lines
	lines := strings.Split(content, "\n")
	var body []string
	for _, line := range lines {
		// Pad or truncate line to fit
		lineW := lipgloss.Width(line)
		if lineW < innerW {
			line = line + strings.Repeat(" ", innerW-lineW)
		}
		body = append(body, "║ "+line+" ║")
	}

	return TerminalDim.Render(top) + "\n" + strings.Join(body, "\n") + "\n" + TerminalDim.Render(bottom)
}

// SparkLine creates an ASCII spark line graph
func SparkLine(values []int, width int) string {
	if len(values) == 0 {
		return strings.Repeat("·", width)
	}
	// Normalize to width
	if len(values) > width {
		// Sample values
		step := len(values) / width
		var sampled []int
		for i := 0; i < width && i*step < len(values); i++ {
			sampled = append(sampled, values[i*step])
		}
		values = sampled
	}

	// Find max
	max := 1
	for _, v := range values {
		if v > max {
			max = v
		}
	}

	// Render sparkline
	sparks := []rune{'·', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	var sb strings.Builder
	for _, v := range values {
		idx := int(float64(v) / float64(max) * float64(len(sparks)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparks) {
			idx = len(sparks) - 1
		}
		sb.WriteRune(sparks[idx])
	}
	// Pad if needed
	for sb.Len() < width {
		sb.WriteRune('·')
	}
	return Terminal.Render(sb.String())
}

// BarGraph creates a horizontal bar
func BarGraph(value, max, width int, label string) string {
	if max <= 0 {
		max = 1
	}
	if width < 10 {
		width = 20
	}
	barW := width - len(label) - 10 // Reserve space for label and value
	if barW < 5 {
		barW = 5
	}
	filled := int(float64(value) / float64(max) * float64(barW))
	if filled > barW {
		filled = barW
	}
	bar := Terminal.Render(strings.Repeat("█", filled)) + Dim.Render(strings.Repeat("░", barW-filled))
	return fmt.Sprintf("%s %s %s", Warn.Render(label), bar, Muted.Render(fmt.Sprintf("%d/%d", value, max)))
}
