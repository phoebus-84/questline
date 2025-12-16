package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"questline/internal/engine"
	"questline/internal/storage"
	"questline/internal/ui"
)

// Panel focus states
type panelFocus int

const (
	focusQuests panelFocus = iota
	focusFocus
)

type boardModel struct {
	ctx context.Context
	svc *engine.Service

	width  int
	height int

	player *storage.Player
	tasks  []storage.Task

	// XP history for graphs
	weeklyXP  []int // XP per day for last 7 days
	monthlyXP []int // XP per week for last 4 weeks

	// Achievements
	achievements []engine.Achievement

	expanded map[int64]bool
	selected int
	focus    panelFocus

	help    help.Model
	keys    keyMap
	spinner spinner.Model

	lastLog   string
	loading   bool
	err       error
	showHelp  bool
	compactUI bool // for small terminals

	// Confirmation state
	confirmDelete bool
	deleteTaskID  int64
}

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Toggle   key.Binding
	Complete key.Binding
	Delete   key.Binding
	Refresh  key.Binding
	Tab      key.Binding
	Help     key.Binding
	Quit     key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Complete, k.Delete, k.Refresh, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Toggle},
		{k.Complete, k.Delete, k.Refresh, k.Tab},
		{k.Help, k.Quit},
	}
}

type loadedMsg struct {
	player       *storage.Player
	tasks        []storage.Task
	weeklyXP     []int
	monthlyXP    []int
	achievements []engine.Achievement
	err          error
}

type completedMsg struct {
	id  int64
	res *engine.CompleteResult
	err error
}

type deletedMsg struct {
	id  int64
	err error
}

func newBoardModel(ctx context.Context, svc *engine.Service) boardModel {
	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	sp.Style = ui.Terminal

	return boardModel{
		ctx:      ctx,
		svc:      svc,
		expanded: map[int64]bool{},
		loading:  true,
		lastLog:  "System initialized.",
		help:     help.New(),
		spinner:  sp,
		focus:    focusQuests,
		keys: keyMap{
			Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
			Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
			Toggle:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("⏎", "expand")),
			Complete: key.NewBinding(key.WithKeys("c", "space"), key.WithHelp("c/␣", "complete")),
			Delete:   key.NewBinding(key.WithKeys("d", "backspace"), key.WithHelp("d/⌫", "delete")),
			Refresh:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
			Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "switch panel")),
			Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
			Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		},
	}
}

func (m boardModel) Init() tea.Cmd {
	return tea.Batch(m.loadCmd(), m.spinner.Tick)
}

func (m boardModel) loadCmd() tea.Cmd {
	return func() tea.Msg {
		p, err := m.svc.PlayerRepo().GetOrCreateMain(m.ctx)
		if err != nil {
			return loadedMsg{err: err}
		}
		tasks, err := m.svc.TaskRepo().ListAll(m.ctx)
		if err != nil {
			return loadedMsg{err: err}
		}

		// Load XP history for graphs
		now := time.Now()
		weeklyXP := make([]int, 7)
		monthlyXP := make([]int, 4)

		// Get XP by day for the last 7 days
		for i := 0; i < 7; i++ {
			day := now.AddDate(0, 0, -6+i)
			dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
			dayEnd := dayStart.AddDate(0, 0, 1).Add(-time.Second)
			xpMap, _ := m.svc.CompletionRepo().XPByDay(m.ctx, dayStart, dayEnd)
			for _, xp := range xpMap {
				weeklyXP[i] += xp
			}
		}

		// Get XP by week for the last 4 weeks
		for i := 0; i < 4; i++ {
			weekStart := now.AddDate(0, 0, -7*(4-i-1)-int(now.Weekday()))
			weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
			weekEnd := weekStart.AddDate(0, 0, 7).Add(-time.Second)
			xpMap, _ := m.svc.CompletionRepo().XPByDay(m.ctx, weekStart, weekEnd)
			for _, xp := range xpMap {
				monthlyXP[i] += xp
			}
		}

		// Load achievements
		achievements, _ := engine.GetAchievementsForPlayer(m.ctx, m.svc)

		return loadedMsg{player: p, tasks: tasks, weeklyXP: weeklyXP, monthlyXP: monthlyXP, achievements: achievements}
	}
}

func (m boardModel) completeCmd(id int64) tea.Cmd {
	return func() tea.Msg {
		res, err := m.svc.CompleteTask(m.ctx, id)
		return completedMsg{id: id, res: res, err: err}
	}
}

func (m boardModel) deleteCmd(id int64) tea.Cmd {
	return func() tea.Msg {
		err := m.svc.TaskRepo().Delete(m.ctx, id)
		return deletedMsg{id: id, err: err}
	}
}

func (m boardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.compactUI = m.width < 80 || m.height < 24
		m.help.Width = m.width
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case loadedMsg:
		m.loading = false
		m.err = msg.err
		if msg.err != nil {
			m.lastLog = "ERROR: " + msg.err.Error()
			return m, nil
		}
		m.player = msg.player
		m.tasks = msg.tasks
		m.weeklyXP = msg.weeklyXP
		m.monthlyXP = msg.monthlyXP
		m.achievements = msg.achievements
		// Default-expand roots that have children.
		children := indexChildren(m.tasks)
		for _, t := range m.tasks {
			if t.ParentID == nil && len(children[t.ID]) > 0 {
				m.expanded[t.ID] = true
			}
		}
		m.lastLog = fmt.Sprintf("Data loaded @ %s", time.Now().Format("15:04:05"))
		return m, nil
	case completedMsg:
		if msg.err != nil {
			m.lastLog = "ERROR: " + msg.err.Error()
			return m, nil
		}
		levelMsg := ""
		if msg.res.LevelAfter > msg.res.LevelBefore {
			levelMsg = fmt.Sprintf(" ▲▲▲ LEVEL UP! %d → %d ▲▲▲", msg.res.LevelBefore, msg.res.LevelAfter)
		}
		m.lastLog = fmt.Sprintf("✓ Task #%d complete: +%d XP%s", msg.res.TaskID, msg.res.XPAwarded, levelMsg)
		return m, m.loadCmd()
	case deletedMsg:
		m.confirmDelete = false
		m.deleteTaskID = 0
		if msg.err != nil {
			m.lastLog = "ERROR: " + msg.err.Error()
			return m, nil
		}
		m.lastLog = fmt.Sprintf("✗ Task #%d deleted", msg.id)
		return m, m.loadCmd()
	case tea.KeyMsg:
		// Handle confirmation mode
		if m.confirmDelete {
			switch msg.String() {
			case "y", "Y":
				m.lastLog = fmt.Sprintf("Deleting task #%d...", m.deleteTaskID)
				return m, m.deleteCmd(m.deleteTaskID)
			case "n", "N", "escape":
				m.confirmDelete = false
				m.deleteTaskID = 0
				m.lastLog = "Delete cancelled"
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "tab":
			if m.focus == focusQuests {
				m.focus = focusFocus
			} else {
				m.focus = focusQuests
			}
			return m, nil
		case "r":
			m.loading = true
			m.lastLog = "Refreshing..."
			return m, m.loadCmd()
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
			return m, nil
		case "down", "j":
			lines := m.questLines()
			if m.selected < len(lines)-1 {
				m.selected++
			}
			return m, nil
		case "enter":
			lines := m.questLines()
			if m.selected < 0 || m.selected >= len(lines) {
				return m, nil
			}
			line := lines[m.selected]
			if line.hasChildren {
				m.expanded[line.id] = !m.expanded[line.id]
				return m, nil
			}
			return m, nil
		case "c", " ":
			lines := m.questLines()
			if m.selected < 0 || m.selected >= len(lines) {
				return m, nil
			}
			line := lines[m.selected]
			if line.isProject {
				m.lastLog = "Cannot complete project directly. Complete its tasks first."
				return m, nil
			}
			// Only complete non-done leaf tasks/habits.
			t := findTask(m.tasks, line.id)
			if t == nil {
				m.lastLog = "Task not found."
				return m, nil
			}
			if t.Status == "done" {
				m.lastLog = "Already completed."
				return m, nil
			}
			m.lastLog = fmt.Sprintf("Completing task #%d...", t.ID)
			return m, m.completeCmd(t.ID)
		case "d", "backspace":
			lines := m.questLines()
			if m.selected < 0 || m.selected >= len(lines) {
				return m, nil
			}
			line := lines[m.selected]
			t := findTask(m.tasks, line.id)
			if t == nil {
				m.lastLog = "Task not found."
				return m, nil
			}
			// Ask for confirmation
			m.confirmDelete = true
			m.deleteTaskID = t.ID
			m.lastLog = fmt.Sprintf("Delete task #%d '%s'? (y/n)", t.ID, truncate(t.Title, 20))
			return m, nil
		}
	}
	return m, nil
}

type questLine struct {
	id          int64
	depth       int
	title       string
	status      string
	isProject   bool
	isHabit     bool
	hasChildren bool
	expanded    bool
}

func (m boardModel) questLines() []questLine {
	if len(m.tasks) == 0 {
		return nil
	}
	children := indexChildren(m.tasks)
	roots := rootIDs(m.tasks)

	var out []questLine
	var walk func(id int64, depth int)
	walk = func(id int64, depth int) {
		t := findTask(m.tasks, id)
		if t == nil {
			return
		}
		kids := children[id]
		q := questLine{
			id:          id,
			depth:       depth,
			title:       t.Title,
			status:      t.Status,
			isProject:   t.IsProject,
			isHabit:     t.IsHabit,
			hasChildren: len(kids) > 0,
			expanded:    m.expanded[id],
		}
		out = append(out, q)
		if len(kids) == 0 {
			return
		}
		if !m.expanded[id] {
			return
		}
		for _, kid := range kids {
			walk(kid, depth+1)
		}
	}

	for _, id := range roots {
		walk(id, 0)
	}
	if m.selected >= len(out) {
		m.selected = len(out) - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
	return out
}

func (m boardModel) View() string {
	if m.err != nil {
		return ui.Bad.Render("█ ERROR █") + "\n\n" + m.err.Error() + "\n\nPress q to quit.\n"
	}

	// Calculate dimensions
	w := m.width
	h := m.height
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 24
	}

	// Header takes ~3 lines, footer ~2 lines
	headerH := 3
	footerH := 2
	bodyH := h - headerH - footerH
	if bodyH < 10 {
		bodyH = 10
	}

	// Render components
	header := m.renderHeader(w)
	footer := m.renderFooter(w)

	// Calculate panel widths
	sidebarW := 28
	if w < 100 {
		sidebarW = 24
	}
	if w < 80 {
		sidebarW = 20
	}
	mainW := w - sidebarW - 4 // borders + padding
	if mainW < 30 {
		mainW = 30
	}

	// Render panels with proper heights
	panelH := bodyH - 2 // border space

	sidebar := m.renderSidebar(sidebarW, panelH)
	main := m.renderMain(mainW, panelH)

	// Style panels with retro borders
	sidebarStyle := lipgloss.NewStyle().
		Width(sidebarW).
		Height(panelH).
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(ui.TerminalDim.GetForeground()).
		Padding(0, 1)

	mainStyle := lipgloss.NewStyle().
		Width(mainW).
		Height(panelH).
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(ui.TerminalDim.GetForeground()).
		Padding(0, 1)

	sidePanel := sidebarStyle.Render(sidebar)
	mainPanel := mainStyle.Render(main)

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidePanel, mainPanel)

	// Full screen layout
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m boardModel) renderHeader(w int) string {
	if m.player == nil {
		title := ui.TerminalBold.Render("▓▓▓ QUESTLINE ▓▓▓")
		return title + " " + ui.Terminal.Render(m.spinner.View()+" LOADING...")
	}

	lvl := engine.LevelForTotalXP(m.player.XPTotal)
	curXP := m.player.XPTotal - engine.XPRequiredForLevel(lvl)
	needXP := engine.XPRequiredForLevel(lvl+1) - engine.XPRequiredForLevel(lvl)

	// Build header line
	title := ui.Gold.Render("▓▓▓ QUESTLINE ▓▓▓")

	// Stats in terminal style
	stats := fmt.Sprintf("%s %s  %s %s  %s %d/%d",
		ui.TerminalDim.Render("LVL"),
		ui.TerminalBold.Render(fmt.Sprintf("%d", lvl)),
		ui.TerminalDim.Render("XP"),
		ui.Terminal.Render(fmt.Sprintf("%d", m.player.XPTotal)),
		ui.TerminalDim.Render("NEXT"),
		curXP,
		needXP,
	)

	// XP bar
	barW := 20
	if w > 100 {
		barW = 30
	}
	bar := progressBarRetro(curXP, needXP, barW)

	// Compose header
	left := title
	right := stats + " " + bar

	gap := w - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	line1 := left + strings.Repeat(" ", gap) + right

	// Separator line
	sep := ui.TerminalDim.Render(strings.Repeat("─", w))

	return line1 + "\n" + sep
}

func (m boardModel) renderSidebar(w, h int) string {
	if m.player == nil {
		return ui.Gold.Render("◆ STATS ◆") + "\n\n" + ui.Terminal.Render(m.spinner.View()+" loading...")
	}

	var lines []string

	// Attributes section
	lines = append(lines, ui.Gold.Render("◆ ATTRIBUTES ◆"))
	lines = append(lines, "")

	barW := w - 14
	if barW < 8 {
		barW = 8
	}

	// All 9 attributes
	attrs := []struct {
		icon string
		name string
		xp   int
	}{
		{"💪", "STR", m.player.XPStr},
		{"🧠", "INT", m.player.XPInt},
		{"🧘", "WIS", m.player.XPWis},
		{"🎨", "ART", m.player.XPArt},
		{"🏠", "HOME", m.player.XPHome},
		{"🌲", "OUT", m.player.XPOut},
		{"📚", "READ", m.player.XPRead},
		{"🎬", "CINE", m.player.XPCinema},
		{"💼", "WORK", m.player.XPCareer},
	}

	for _, a := range attrs {
		lines = append(lines, renderAttrRetro(a.icon, a.name, a.xp, barW))
	}

	// Stats section
	lines = append(lines, "")
	lines = append(lines, ui.Gold.Render("◆ STATS ◆"))
	lines = append(lines, "")

	// Count tasks
	pending, done, habits := 0, 0, 0
	for _, t := range m.tasks {
		if t.IsHabit {
			habits++
		}
		switch t.Status {
		case "pending", "active", "planning":
			pending++
		case "done":
			done++
		}
	}
	lines = append(lines, fmt.Sprintf("%s %d", ui.TerminalDim.Render("Active:"), pending))
	lines = append(lines, fmt.Sprintf("%s %d", ui.TerminalDim.Render("Done:  "), done))
	lines = append(lines, fmt.Sprintf("%s %d", ui.TerminalDim.Render("Habits:"), habits))

	// XP Graph section (weekly)
	lines = append(lines, "")
	lines = append(lines, ui.Gold.Render("◆ XP (7 DAYS) ◆"))
	lines = append(lines, renderSparkGraph(m.weeklyXP, barW))

	// Total XP this week
	weekTotal := 0
	for _, xp := range m.weeklyXP {
		weekTotal += xp
	}
	lines = append(lines, ui.TerminalDim.Render(fmt.Sprintf("Total: %d XP", weekTotal)))

	// Achievements section
	lines = append(lines, "")
	lines = append(lines, ui.Gold.Render("◆ BADGES ◆"))
	earnedCount := 0
	totalCount := len(m.achievements)
	recentBadges := ""
	for _, a := range m.achievements {
		if a.Earned {
			earnedCount++
			if len(recentBadges) < barW {
				recentBadges += a.Icon
			}
		}
	}
	if recentBadges == "" {
		recentBadges = ui.TerminalDim.Render("(none yet)")
	}
	lines = append(lines, recentBadges)
	lines = append(lines, ui.TerminalDim.Render(fmt.Sprintf("%d/%d earned", earnedCount, totalCount)))

	// Keys help
	lines = append(lines, "")
	lines = append(lines, ui.Gold.Render("◆ KEYS ◆"))
	if m.showHelp {
		lines = append(lines, ui.TerminalDim.Render("↑↓/jk  navigate"))
		lines = append(lines, ui.TerminalDim.Render("⏎      expand"))
		lines = append(lines, ui.TerminalDim.Render("c/␣    complete"))
		lines = append(lines, ui.TerminalDim.Render("d/⌫    delete"))
		lines = append(lines, ui.TerminalDim.Render("r      refresh"))
		lines = append(lines, ui.TerminalDim.Render("?      toggle help"))
		lines = append(lines, ui.TerminalDim.Render("q      quit"))
	} else {
		lines = append(lines, ui.TerminalDim.Render("? for help"))
	}

	return strings.Join(lines, "\n")
}

func (m boardModel) renderMain(w, h int) string {
	if m.loading {
		return ui.Gold.Render("◆ LOADING ◆") + "\n\n" + ui.Terminal.Render(m.spinner.View()+" fetching data...")
	}

	var out []string

	// Focus section (top tasks)
	out = append(out, ui.Gold.Render("◆ FOCUS ◆"))
	focus := m.focusTasks(3)
	if len(focus) == 0 {
		out = append(out, ui.TerminalDim.Render("  (no pending tasks)"))
	} else {
		for _, t := range focus {
			icon := kindIconRetro(t.IsProject, t.IsHabit)
			xpStr := ui.TerminalDim.Render(fmt.Sprintf("+%dXP", t.XPValue))
			out = append(out, fmt.Sprintf("  %s #%d %s %s", icon, t.ID, truncate(t.Title, w-20), xpStr))
		}
	}

	out = append(out, "")
	out = append(out, ui.Gold.Render("◆ QUEST LOG ◆"))

	lines := m.questLines()
	if len(lines) == 0 {
		out = append(out, ui.TerminalDim.Render("  (empty)"))
		return strings.Join(out, "\n")
	}

	// Calculate visible lines based on height
	maxVisible := h - len(out) - 2
	if maxVisible < 5 {
		maxVisible = 5
	}

	// Scrolling: ensure selected is visible
	startIdx := 0
	if m.selected >= maxVisible {
		startIdx = m.selected - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	// Show scroll indicator if needed
	if startIdx > 0 {
		out = append(out, ui.TerminalDim.Render("  ▲ more above ▲"))
	}

	for i := startIdx; i < endIdx; i++ {
		ql := lines[i]
		indent := strings.Repeat("  ", ql.depth)
		icon := kindIconRetro(ql.isProject, ql.isHabit)

		fold := "  "
		if ql.hasChildren {
			if ql.expanded {
				fold = "▾ "
			} else {
				fold = "▸ "
			}
		}

		statusIcon := statusIconRetro(ql.status)
		title := truncate(ql.title, w-ql.depth*2-15)

		row := fmt.Sprintf("%s%s%s %s %s", indent, fold, icon, title, statusIcon)

		if i == m.selected {
			// Highlight selected row
			row = ui.SelectedRow.Render(row)
		}
		out = append(out, row)
	}

	if endIdx < len(lines) {
		out = append(out, ui.TerminalDim.Render("  ▼ more below ▼"))
	}

	return strings.Join(out, "\n")
}

func (m boardModel) renderFooter(w int) string {
	sep := ui.TerminalDim.Render(strings.Repeat("─", w))

	// Status line in terminal style
	var status string
	if m.loading {
		status = ui.Terminal.Render(m.spinner.View()+" ") + ui.TerminalDim.Render(m.lastLog)
	} else {
		status = ui.Terminal.Render("> ") + ui.TerminalDim.Render(m.lastLog)
	}

	return sep + "\n" + status
}

func (m boardModel) focusTasks(n int) []storage.Task {
	var leaf []storage.Task
	children := indexChildren(m.tasks)
	for _, t := range m.tasks {
		if t.IsProject {
			continue
		}
		if t.Status == "done" {
			continue
		}
		if len(children[t.ID]) > 0 {
			continue
		}
		switch t.Status {
		case "pending", "active":
			leaf = append(leaf, t)
		}
	}
	sort.Slice(leaf, func(i, j int) bool {
		// Prefer due soon, then ID.
		ai := leaf[i].DueDate
		aj := leaf[j].DueDate
		if ai == nil && aj != nil {
			return false
		}
		if ai != nil && aj == nil {
			return true
		}
		if ai != nil && aj != nil {
			if !ai.Equal(*aj) {
				return ai.Before(*aj)
			}
		}
		return leaf[i].ID < leaf[j].ID
	})
	if len(leaf) > n {
		leaf = leaf[:n]
	}
	return leaf
}

func renderAttr(label string, xp int) string {
	lvl := engine.AttributeLevelForXP(xp)
	cur := engine.XPRequiredForLevel(lvl)
	next := engine.XPRequiredForLevel(lvl + 1)
	bar := progressBarStyled(xp-cur, next-cur, 14)
	return fmt.Sprintf("%s %s %s", label, ui.Muted.Render(fmt.Sprintf("L%d", lvl)), bar)
}

func renderAttrRetro(icon, name string, xp, barW int) string {
	lvl := engine.AttributeLevelForXP(xp)
	cur := engine.XPRequiredForLevel(lvl)
	next := engine.XPRequiredForLevel(lvl + 1)
	bar := progressBarRetro(xp-cur, next-cur, barW)
	return fmt.Sprintf("%s %s %s %s", icon, ui.TerminalDim.Render(fmt.Sprintf("%-4s", name)), ui.Terminal.Render(fmt.Sprintf("L%d", lvl)), bar)
}

func progressBarStyled(value int, total int, width int) string {
	if total <= 0 {
		total = 1
	}
	if width <= 3 {
		width = 3
	}
	if value < 0 {
		value = 0
	}
	if value > total {
		value = total
	}
	ratio := float64(value) / float64(total)
	filled := int(ratio * float64(width))
	if filled > width {
		filled = width
	}
	fill := strings.Repeat("█", filled)
	empty := strings.Repeat("░", width-filled)
	return ui.Muted.Render("[") + ui.Good.Render(fill) + ui.Dim.Render(empty) + ui.Muted.Render("]")
}

func progressBarRetro(value int, total int, width int) string {
	if total <= 0 {
		total = 1
	}
	if width <= 2 {
		width = 2
	}
	if value < 0 {
		value = 0
	}
	if value > total {
		value = total
	}
	ratio := float64(value) / float64(total)
	filled := int(ratio * float64(width))
	if filled > width {
		filled = width
	}
	fill := strings.Repeat("▓", filled)
	empty := strings.Repeat("░", width-filled)
	return ui.Terminal.Render(fill) + ui.TerminalDim.Render(empty)
}

func kindIconRetro(isProject, isHabit bool) string {
	if isProject {
		return ui.Gold.Render("◈")
	}
	if isHabit {
		return ui.Terminal.Render("○")
	}
	return ui.TerminalDim.Render("•")
}

func statusIconRetro(status string) string {
	switch status {
	case "done":
		return ui.Terminal.Render("✓")
	case "active":
		return ui.Gold.Render("►")
	case "planning":
		return ui.TerminalDim.Render("…")
	default:
		return ui.TerminalDim.Render("·")
	}
}

func truncate(s string, maxLen int) string {
	if maxLen < 4 {
		maxLen = 4
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// renderSparkGraph renders a spark line graph using block characters
func renderSparkGraph(values []int, width int) string {
	if len(values) == 0 {
		return ui.TerminalDim.Render("(no data)")
	}

	// Find max value for scaling
	maxVal := 1
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}

	// Spark characters from low to high
	sparks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	var result strings.Builder
	for _, v := range values {
		if v == 0 {
			result.WriteString(ui.TerminalDim.Render("░"))
		} else {
			// Scale value to spark index (0-7)
			idx := (v * 7) / maxVal
			if idx > 7 {
				idx = 7
			}
			result.WriteString(ui.Terminal.Render(string(sparks[idx])))
		}
	}
	return result.String()
}

// renderBarGraph renders a horizontal bar graph
func renderBarGraph(values []int, labels []string, width int) []string {
	if len(values) == 0 {
		return []string{ui.TerminalDim.Render("(no data)")}
	}

	maxVal := 1
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}

	barWidth := width - 8 // Leave room for label
	if barWidth < 5 {
		barWidth = 5
	}

	var lines []string
	for i, v := range values {
		label := ""
		if i < len(labels) {
			label = labels[i]
		}

		filled := 0
		if maxVal > 0 {
			filled = (v * barWidth) / maxVal
		}

		bar := strings.Repeat("▓", filled) + strings.Repeat("░", barWidth-filled)
		lines = append(lines, fmt.Sprintf("%-3s %s", label, ui.Terminal.Render(bar)))
	}
	return lines
}

func findTask(tasks []storage.Task, id int64) *storage.Task {
	for i := range tasks {
		if tasks[i].ID == id {
			return &tasks[i]
		}
	}
	return nil
}

func rootIDs(tasks []storage.Task) []int64 {
	var roots []int64
	for _, t := range tasks {
		if t.ParentID == nil {
			roots = append(roots, t.ID)
		}
	}
	sort.Slice(roots, func(i, j int) bool { return roots[i] < roots[j] })
	return roots
}

func indexChildren(tasks []storage.Task) map[int64][]int64 {
	children := map[int64][]int64{}
	for _, t := range tasks {
		if t.ParentID == nil {
			continue
		}
		children[*t.ParentID] = append(children[*t.ParentID], t.ID)
	}
	for k := range children {
		sort.Slice(children[k], func(i, j int) bool { return children[k][i] < children[k][j] })
	}
	return children
}
