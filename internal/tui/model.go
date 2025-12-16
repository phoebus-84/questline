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

type boardModel struct {
	ctx context.Context
	svc *engine.Service

	width  int
	height int

	player *storage.Player
	tasks  []storage.Task

	expanded map[int64]bool
	selected int

	help    help.Model
	keys    keyMap
	spinner spinner.Model

	lastLog string
	loading bool
	err     error
}

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Toggle   key.Binding
	Complete key.Binding
	Refresh  key.Binding
	Quit     key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Toggle, k.Complete, k.Refresh, k.Quit}
}

type loadedMsg struct {
	player *storage.Player
	tasks  []storage.Task
	err    error
}

type completedMsg struct {
	id  int64
	res *engine.CompleteResult
	err error
}

func newBoardModel(ctx context.Context, svc *engine.Service) boardModel {
	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	sp.Style = ui.Muted

	return boardModel{
		ctx:      ctx,
		svc:      svc,
		expanded: map[int64]bool{},
		loading:  true,
		lastLog:  "Loaded.",
		help:     help.New(),
		spinner:  sp,
		keys: keyMap{
			Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move")),
			Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move")),
			Toggle:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "expand")),
			Complete: key.NewBinding(key.WithKeys("c", "space"), key.WithHelp("c/space", "complete")),
			Refresh:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
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
		return loadedMsg{player: p, tasks: tasks}
	}
}

func (m boardModel) completeCmd(id int64) tea.Cmd {
	return func() tea.Msg {
		res, err := m.svc.CompleteTask(m.ctx, id)
		return completedMsg{id: id, res: res, err: err}
	}
}

func (m boardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case loadedMsg:
		m.loading = false
		m.err = msg.err
		if msg.err != nil {
			m.lastLog = "Load failed: " + msg.err.Error()
			return m, nil
		}
		m.player = msg.player
		m.tasks = msg.tasks
		// Default-expand roots that have children.
		children := indexChildren(m.tasks)
		for _, t := range m.tasks {
			if t.ParentID == nil && len(children[t.ID]) > 0 {
				m.expanded[t.ID] = true
			}
		}
		m.lastLog = fmt.Sprintf("Refreshed at %s.", time.Now().Format("15:04:05"))
		return m, nil
	case completedMsg:
		if msg.err != nil {
			m.lastLog = "Complete failed: " + msg.err.Error()
			return m, nil
		}
		m.lastLog = fmt.Sprintf("Completed %d: +%d XP (level %d → %d)", msg.res.TaskID, msg.res.XPAwarded, msg.res.LevelBefore, msg.res.LevelAfter)
		return m, m.loadCmd()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			m.loading = true
			m.lastLog = "Refreshing…"
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
				m.lastLog = "Select a leaf task/habit to complete."
				return m, nil
			}
			// Only complete non-done leaf tasks/habits.
			t := findTask(m.tasks, line.id)
			if t == nil {
				m.lastLog = "Task not found."
				return m, nil
			}
			if t.Status == "done" {
				m.lastLog = "Already done."
				return m, nil
			}
			m.lastLog = fmt.Sprintf("Completing %d…", t.ID)
			return m, m.completeCmd(t.ID)
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
		return ui.Bad.Render(ui.IconError+" Error") + ": " + m.err.Error() + "\n\nPress q to quit.\n"
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	leftW := 30
	if m.width > 0 {
		maxLeft := (m.width - 2) / 2
		if maxLeft < leftW {
			leftW = maxLeft
		}
		if leftW < 22 {
			leftW = 22
		}
	}

	rightW := 0
	if m.width > 0 {
		rightW = m.width - leftW - 2
		if rightW < 20 {
			rightW = 20
		}
	}

	left := ui.Panel.Width(leftW).Render(m.renderSidebar())
	right := ui.Panel.Render(m.renderMain())
	if rightW > 0 {
		right = ui.Panel.Width(rightW).Render(m.renderMain())
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	return header + "\n" + body + "\n" + footer
}

func (m boardModel) renderHeader() string {
	if m.player == nil {
		return ui.Heading(ui.IconQuest, "Questline") + " " + ui.Muted.Render(m.spinner.View()+" loading…")
	}
	lvl := engine.LevelForTotalXP(m.player.XPTotal)
	bar := progressBarStyled(
		m.player.XPTotal-engine.XPRequiredForLevel(lvl),
		engine.XPRequiredForLevel(lvl+1)-engine.XPRequiredForLevel(lvl),
		30,
	)
	left := ui.Heading(ui.IconQuest, "Questline")
	right := fmt.Sprintf("%s %s  %s %s  %s %d %s",
		ui.Muted.Render("Player"), ui.Key.Render(m.player.Key),
		ui.Muted.Render("Level"), ui.Key.Render(fmt.Sprintf("%d", lvl)),
		ui.Muted.Render("XP"), m.player.XPTotal, bar,
	)
	if m.width > 0 {
		gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
		if gap > 1 {
			return left + strings.Repeat(" ", gap) + right
		}
	}
	return left + "  " + right
}

func (m boardModel) renderSidebar() string {
	if m.player == nil {
		return ui.PanelTitle.Render("📊 Stats") + "\n\n" + ui.Muted.Render(m.spinner.View()+" loading…")
	}
	lines := []string{ui.PanelTitle.Render("📊 Attributes")}
	// Original 4 attributes
	lines = append(lines, renderAttr("💪 STR", m.player.XPStr))
	lines = append(lines, renderAttr("🧠 INT", m.player.XPInt))
	lines = append(lines, renderAttr("🧘 WIS", m.player.XPWis))
	lines = append(lines, renderAttr("🎨 ART", m.player.XPArt))
	// New 5 attributes
	lines = append(lines, renderAttr("🏠 HOME", m.player.XPHome))
	lines = append(lines, renderAttr("🌲 OUT", m.player.XPOut))
	lines = append(lines, renderAttr("📚 READ", m.player.XPRead))
	lines = append(lines, renderAttr("🎬 CINEMA", m.player.XPCinema))
	lines = append(lines, renderAttr("💼 CAREER", m.player.XPCareer))
	lines = append(lines, "")
	lines = append(lines, ui.PanelTitle.Render("⌨️  Keys"))
	lines = append(lines, ui.Muted.Render(m.help.ShortHelpView(m.keys.ShortHelp())))
	return strings.Join(lines, "\n")
}

func (m boardModel) renderMain() string {
	if m.loading {
		return ui.PanelTitle.Render("✨ Loading") + "\n\n" + ui.Muted.Render(m.spinner.View()+" fetching quests…")
	}
	focus := m.focusTasks(3)
	var out []string
	out = append(out, ui.PanelTitle.Render("🎯 Focus"))
	if len(focus) == 0 {
		out = append(out, ui.Muted.Render("(no pending leaf tasks)"))
	} else {
		for _, t := range focus {
			icon := ui.KindIcon(t.IsProject, t.IsHabit)
			out = append(out, fmt.Sprintf("%s #%d %s %s", icon, t.ID, t.Title, ui.Muted.Render(fmt.Sprintf("(+%d XP)", t.XPValue))))
		}
	}
	out = append(out, "")
	out = append(out, ui.PanelTitle.Render("🗺️ Quest Log"))

	lines := m.questLines()
	if len(lines) == 0 {
		out = append(out, "(empty)")
		return strings.Join(out, "\n")
	}
	for i, ql := range lines {
		indent := strings.Repeat("  ", ql.depth)
		kind := ui.KindIcon(ql.isProject, ql.isHabit)
		fold := "  "
		if ql.hasChildren {
			if ql.expanded {
				fold = "▾ "
			} else {
				fold = "▸ "
			}
		}
		status := ui.StatusText(ql.status)
		row := fmt.Sprintf("%s%s%s %s %s", indent, fold, kind, ql.title, ui.Muted.Render("("+status+")"))
		if i == m.selected {
			row = ui.SelectedRow.Render(row)
		}
		out = append(out, row)
	}
	return strings.Join(out, "\n")
}

func (m boardModel) renderFooter() string {
	line := ui.Muted.Render(m.lastLog)
	if m.loading {
		line = ui.Muted.Render(m.spinner.View()+" ") + line
	}
	return line
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
