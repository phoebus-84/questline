package tui

import (
	"context"
	"io"

	tea "github.com/charmbracelet/bubbletea"

	"questline/internal/engine"
)

func RunBoard(ctx context.Context, svc *engine.Service, out io.Writer) error {
	m := newBoardModel(ctx, svc)
	p := tea.NewProgram(m, tea.WithOutput(out))
	_, err := p.Run()
	return err
}