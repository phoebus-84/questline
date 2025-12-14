package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"questline/internal/engine"
	"questline/internal/storage"
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

	lastLog string
	loading bool
	err     error
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
	return boardModel{
		ctx:      ctx,
		svc:      svc,
		expanded: map[int64]bool{},
		loading:  true,
		lastLog:  "Loaded.",
	}
}

func (m boardModel) Init() tea.Cmd {
	return m.loadCmd()
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
		return "Error: " + m.err.Error() + "\n\nPress q to quit.\n"
	}

	header := m.renderHeader()
	sidebar := m.renderSidebar()
	main := m.renderMain()
	footer := m.renderFooter()

	// Simple 2-column layout.
	leftW := 26
	if m.width > 0 {
		maxLeft := m.width / 2
		if maxLeft < leftW {
			leftW = maxLeft
		}
		if leftW < 18 {
			leftW = 18
		}
	}

	linesLeft := strings.Split(sidebar, "\n")
	linesRight := strings.Split(main, "\n")
	max := len(linesLeft)
	if len(linesRight) > max {
		max = len(linesRight)
	}

	var body strings.Builder
	for i := 0; i < max; i++ {
		l := ""
		r := ""
		if i < len(linesLeft) {
			l = linesLeft[i]
		}
		if i < len(linesRight) {
			r = linesRight[i]
		}
		body.WriteString(padRight(l, leftW))
		body.WriteString("  ")
		body.WriteString(r)
		body.WriteString("\n")
	}

	return header + "\n" + body.String() + footer
}

func (m boardModel) renderHeader() string {
	if m.player == nil {
		return "Questline — loading…"
	}
	lvl := engine.LevelForTotalXP(m.player.XPTotal)
	bar := progressBar(
		m.player.XPTotal-engine.XPRequiredForLevel(lvl),
		engine.XPRequiredForLevel(lvl+1)-engine.XPRequiredForLevel(lvl),
		30,
	)
	return fmt.Sprintf("Questline | Player: %s | Level %d | XP %d %s", m.player.Key, lvl, m.player.XPTotal, bar)
}

func (m boardModel) renderSidebar() string {
	if m.player == nil {
		return "Stats\n\nLoading…"
	}
	lines := []string{"Attributes"}
	lines = append(lines, renderAttr("STR", m.player.XPStr))
	lines = append(lines, renderAttr("INT", m.player.XPInt))
	lines = append(lines, renderAttr("ART", m.player.XPArt))
	lines = append(lines, renderAttr("WIS", m.player.XPWis))
	lines = append(lines, "")
	lines = append(lines, "Keys")
	lines = append(lines, "- ↑/↓ or j/k: move")
	lines = append(lines, "- enter: expand/collapse")
	lines = append(lines, "- c/space: complete")
	lines = append(lines, "- r: refresh")
	lines = append(lines, "- q: quit")
	return strings.Join(lines, "\n")
}

func (m boardModel) renderMain() string {
	if m.loading {
		return "Loading…"
	}
	focus := m.focusTasks(3)
	var out []string
	out = append(out, "Focus")
	if len(focus) == 0 {
		out = append(out, "(no pending leaf tasks)")
	} else {
		for _, t := range focus {
			out = append(out, fmt.Sprintf("- %d %s (xp=%d)", t.ID, t.Title, t.XPValue))
		}
	}
	out = append(out, "")
	out = append(out, "Quest Log")

	lines := m.questLines()
	if len(lines) == 0 {
		out = append(out, "(empty)")
		return strings.Join(out, "\n")
	}
	for i, ql := range lines {
		cursor := "  "
		if i == m.selected {
			cursor = "> "
		}
		indent := strings.Repeat("  ", ql.depth)
		icon := ""
		if ql.isProject {
			icon = "[P] "
		} else if ql.isHabit {
			icon = "[H] "
		}
		fold := "  "
		if ql.hasChildren {
			if ql.expanded {
				fold = "▾ "
			} else {
				fold = "▸ "
			}
		}
		out = append(out, fmt.Sprintf("%s%s%s%s%s (status=%s)", cursor, indent, fold, icon, ql.title, ql.status))
	}
	return strings.Join(out, "\n")
}

func (m boardModel) renderFooter() string {
	return "\n" + m.lastLog
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
	bar := progressBar(xp-cur, next-cur, 14)
	return fmt.Sprintf("- %s L%d %s", label, lvl, bar)
}

func progressBar(value int, total int, width int) string {
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
	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", width-filled) + "]"
}

func padRight(s string, width int) string {
	if width <= 0 {
		return s
	}
	r := []rune(s)
	if len(r) >= width {
		return string(r[:width])
	}
	return s + strings.Repeat(" ", width-len(r))
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
