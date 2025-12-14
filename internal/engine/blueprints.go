package engine

import (
	"context"
	"fmt"
	"strings"

	"questline/internal/storage"
)

type BlueprintStatus string

const (
	BlueprintLocked    BlueprintStatus = "locked"
	BlueprintAvailable BlueprintStatus = "available"
	BlueprintActive    BlueprintStatus = "active"
	BlueprintCompleted BlueprintStatus = "completed"
)

type BlueprintKind string

const (
	BlueprintKindTask    BlueprintKind = "task"
	BlueprintKindProject BlueprintKind = "project"
	BlueprintKindHabit   BlueprintKind = "habit"
)

type BlueprintDef struct {
	Code string
	Kind BlueprintKind

	Title      string
	Difficulty Difficulty
	Attribute  Attribute
	HabitEvery HabitInterval

	Unlock func(ctx context.Context, svc *Service, p *storage.Player) (bool, error)
}

func builtinBlueprints() []BlueprintDef {
	return []BlueprintDef{
		{
			Code:       "str_starter",
			Kind:       BlueprintKindHabit,
			Title:      "Push-ups",
			Difficulty: DifficultyEasy,
			Attribute:  AttributeSTR,
			HabitEvery: HabitIntervalDaily,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelHabits, nil
			},
		},
		{
			Code:      "art_reader",
			Kind:      BlueprintKindProject,
			Title:     "Read a Book",
			Attribute: AttributeART,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				artLevel := AttributeLevelForXP(p.XPArt)
				return p.Level >= LevelProjects && artLevel >= 1, nil
			},
		},
		{
			Code:       "art_critic",
			Kind:       BlueprintKindTask,
			Title:      "Write a short review",
			Difficulty: DifficultyMedium,
			Attribute:  AttributeART,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				artLevel := AttributeLevelForXP(p.XPArt)
				if p.Level < LevelProjects || artLevel < 2 {
					return false, nil
				}
				has, err := svc.tasks.HasCompletedProjectTitle(ctx, "Read a Book")
				if err != nil {
					return false, err
				}
				return has, nil
			},
		},
	}
}

func normalizeBlueprintCode(code string) (string, error) {
	c := strings.TrimSpace(strings.ToLower(code))
	if c == "" {
		return "", fmt.Errorf("blueprint code is required")
	}
	return c, nil
}

// EvaluateBlueprintUnlocks ensures built-in blueprint rows exist and transitions any
// newly unlocked ones from locked -> available.
func (s *Service) EvaluateBlueprintUnlocks(ctx context.Context) ([]storage.Blueprint, error) {
	p, err := s.getPlayer(ctx)
	if err != nil {
		return nil, err
	}

	defs := builtinBlueprints()
	var newlyAvailable []storage.Blueprint

	for _, def := range defs {
		row, err := s.blueprints.Get(ctx, def.Code)
		if err != nil {
			return nil, err
		}
		if row == nil {
			if err := s.blueprints.Upsert(ctx, storage.Blueprint{Code: def.Code, Status: string(BlueprintLocked)}); err != nil {
				return nil, err
			}
			row = &storage.Blueprint{Code: def.Code, Status: string(BlueprintLocked)}
		}
		if row.Status != string(BlueprintLocked) {
			continue
		}

		ok, err := def.Unlock(ctx, s, p)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		b := storage.Blueprint{Code: def.Code, Status: string(BlueprintAvailable)}
		if err := s.blueprints.Upsert(ctx, b); err != nil {
			return nil, err
		}
		newlyAvailable = append(newlyAvailable, b)
	}

	return newlyAvailable, nil
}

// AcceptBlueprint marks an available blueprint as active and instantiates its
// corresponding task/project/habit.
func (s *Service) AcceptBlueprint(ctx context.Context, code string) (*CreateResult, error) {
	c, err := normalizeBlueprintCode(code)
	if err != nil {
		return nil, err
	}

	// Make sure unlock state is up to date (important if player imported/migrated state).
	if _, err := s.EvaluateBlueprintUnlocks(ctx); err != nil {
		return nil, err
	}

	b, err := s.blueprints.Get(ctx, c)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, fmt.Errorf("unknown blueprint: %s", c)
	}
	if b.Status != string(BlueprintAvailable) {
		return nil, fmt.Errorf("blueprint %s is not available (status=%s)", c, b.Status)
	}

	defs := builtinBlueprints()
	var def *BlueprintDef
	for i := range defs {
		if defs[i].Code == c {
			def = &defs[i]
			break
		}
	}
	if def == nil {
		return nil, fmt.Errorf("unknown blueprint: %s", c)
	}

	var res *CreateResult
	switch def.Kind {
	case BlueprintKindProject:
		res, err = s.CreateProject(ctx, CreateProjectInput{Title: def.Title, Attribute: def.Attribute})
	case BlueprintKindHabit:
		res, err = s.CreateTask(ctx, CreateTaskInput{Title: def.Title, Difficulty: def.Difficulty, Attribute: def.Attribute, IsHabit: true, HabitInterval: def.HabitEvery})
	case BlueprintKindTask:
		res, err = s.CreateTask(ctx, CreateTaskInput{Title: def.Title, Difficulty: def.Difficulty, Attribute: def.Attribute, IsHabit: false})
	default:
		return nil, fmt.Errorf("invalid blueprint kind: %s", def.Kind)
	}
	if err != nil {
		return nil, err
	}

	if err := s.blueprints.Upsert(ctx, storage.Blueprint{Code: c, Status: string(BlueprintActive)}); err != nil {
		return nil, err
	}

	return res, nil
}
