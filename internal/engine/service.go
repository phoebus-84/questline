package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"questline/internal/storage"
)

type Service struct {
	db          *sql.DB
	players     *storage.PlayerRepo
	tasks       *storage.TaskRepo
	completions *storage.CompletionRepo
}

func NewService(db *sql.DB) *Service {
	return &Service{
		db:          db,
		players:     storage.NewPlayerRepo(db),
		tasks:       storage.NewTaskRepo(db),
		completions: storage.NewCompletionRepo(db),
	}
}

func (s *Service) PlayerRepo() *storage.PlayerRepo         { return s.players }
func (s *Service) TaskRepo() *storage.TaskRepo             { return s.tasks }
func (s *Service) CompletionRepo() *storage.CompletionRepo { return s.completions }

func normalizeTitle(title string) (string, error) {
	t := strings.TrimSpace(title)
	if t == "" {
		return "", errors.New("title is required")
	}
	return t, nil
}

func (s *Service) getPlayer(ctx context.Context) (*storage.Player, error) {
	p, err := s.players.GetOrCreateMain(ctx)
	if err != nil {
		return nil, err
	}
	computed := LevelForTotalXP(p.XPTotal)
	if p.Level != computed {
		p.Level = computed
		if err := s.players.Update(ctx, p); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func playerXPForAttribute(p *storage.Player, attr Attribute) int {
	switch attr {
	case AttributeSTR:
		return p.XPStr
	case AttributeINT:
		return p.XPInt
	case AttributeART:
		return p.XPArt
	case AttributeWIS:
		fallthrough
	default:
		return p.XPWis
	}
}

func (s *Service) countActiveLeafTasks(ctx context.Context) (int, error) {
	all, err := s.tasks.ListAll(ctx)
	if err != nil {
		return 0, err
	}

	active := 0
	for _, t := range all {
		if t.IsProject {
			continue
		}
		switch t.Status {
		case "pending", "active":
			active++
		}
	}
	return active, nil
}

func (s *Service) taskDepthFromRoot(ctx context.Context, id int64) (int, error) {
	depth := 0
	cur := id
	seen := 0
	for {
		seen++
		if seen > 10_000 {
			return 0, fmt.Errorf("task parent chain too deep (cycle?)")
		}
		t, err := s.tasks.Get(ctx, cur)
		if err != nil {
			return 0, err
		}
		if t == nil {
			return 0, fmt.Errorf("task %d not found", cur)
		}
		if t.ParentID == nil {
			return depth, nil
		}
		depth++
		cur = *t.ParentID
	}
}
