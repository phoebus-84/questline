package engine

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"questline/internal/storage"
)

type CompleteResult struct {
	TaskID        int64
	XPAwarded     int
	LevelBefore   int
	LevelAfter    int
	LevelUp       bool
	ProjectBonus  bool
	ProjectVolume int
}

func parseStoredAttribute(s string) Attribute {
	s = strings.TrimSpace(strings.ToUpper(s))
	a := Attribute(s)
	if a.IsValid() {
		return a
	}
	return DefaultAttribute
}

// addAttributeXP adds XP to the appropriate attribute on the player.
func addAttributeXP(p *storage.Player, attr Attribute, xp int) {
	switch attr {
	case AttributeSTR:
		p.XPStr += xp
	case AttributeINT:
		p.XPInt += xp
	case AttributeART:
		p.XPArt += xp
	case AttributeHOME:
		p.XPHome += xp
	case AttributeOUT:
		p.XPOut += xp
	case AttributeREAD:
		p.XPRead += xp
	case AttributeCINEMA:
		p.XPCinema += xp
	case AttributeCAREER:
		p.XPCareer += xp
	case AttributeWIS:
		fallthrough
	default:
		p.XPWis += xp
	}
}

func (s *Service) CompleteTask(ctx context.Context, id int64) (*CompleteResult, error) {
	p, err := s.getPlayer(ctx)
	if err != nil {
		return nil, err
	}
	levelBefore := p.Level

	task, err := s.tasks.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, fmt.Errorf("task %d not found", id)
	}
	if task.Status == "done" {
		return nil, fmt.Errorf("task %d is already done", id)
	}

	now := time.Now().UTC()

	if task.IsHabit {
		children, err := s.tasks.ListChildren(ctx, id)
		if err != nil {
			return nil, err
		}
		if len(children) > 0 {
			return nil, fmt.Errorf("habit %d must be a leaf", id)
		}
		if task.HabitInterval == nil {
			return nil, fmt.Errorf("habit %d is missing interval", id)
		}

		interval, err := ParseHabitInterval(*task.HabitInterval)
		if err != nil {
			return nil, err
		}

		since := now.Add(-7 * 24 * time.Hour)
		recentSameDifficulty, err := s.completions.CountSinceWithDifficulty(ctx, id, since, task.Difficulty)
		if err != nil {
			return nil, err
		}
		xp := task.XPValue
		if recentSameDifficulty >= 5 {
			xp = int(math.Round(float64(xp) * 0.50))
		}
		if xp < 1 {
			xp = 1
		}

		attr := parseStoredAttribute(task.Attribute)
		nextDue, err := NextDueDate(now, interval)
		if err != nil {
			return nil, err
		}
		if err := s.tasks.UpdateHabitAfterCompletion(ctx, id, now, nextDue); err != nil {
			return nil, err
		}

		p.XPTotal += xp
		addAttributeXP(p, attr, xp)
		p.Level = LevelForTotalXP(p.XPTotal)
		if err := s.players.Update(ctx, p); err != nil {
			return nil, err
		}
		if _, err := s.completions.Insert(ctx, id, now, task.Difficulty, xp); err != nil {
			return nil, err
		}

		levelUp := p.Level > levelBefore
		if levelUp {
			_, _ = s.EvaluateBlueprintUnlocks(ctx)
		}

		return &CompleteResult{
			TaskID:      id,
			XPAwarded:   xp,
			LevelBefore: levelBefore,
			LevelAfter:  p.Level,
			LevelUp:     levelUp,
		}, nil
	}

	if task.IsProject {
		if task.Status == "planning" {
			return nil, fmt.Errorf("project %d is still planning; add a child task first", id)
		}

		volume, hasUndone, err := s.projectVolumeAndUndone(ctx, id)
		if err != nil {
			return nil, err
		}
		if hasUndone {
			return nil, fmt.Errorf("project %d has unfinished tasks", id)
		}

		bonus := int(math.Round(float64(volume) * 0.10))
		attr := parseStoredAttribute(task.Attribute)

		if err := s.tasks.MarkDone(ctx, id, now); err != nil {
			return nil, err
		}

		p.XPTotal += bonus
		addAttributeXP(p, attr, bonus)
		p.Level = LevelForTotalXP(p.XPTotal)
		if err := s.players.Update(ctx, p); err != nil {
			return nil, err
		}
		if _, err := s.completions.Insert(ctx, id, now, task.Difficulty, bonus); err != nil {
			return nil, err
		}

		levelUp := p.Level > levelBefore
		if levelUp {
			_, _ = s.EvaluateBlueprintUnlocks(ctx)
		}

		return &CompleteResult{
			TaskID:        id,
			XPAwarded:     bonus,
			LevelBefore:   levelBefore,
			LevelAfter:    p.Level,
			LevelUp:       levelUp,
			ProjectBonus:  true,
			ProjectVolume: volume,
		}, nil
	}

	children, err := s.tasks.ListChildren(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(children) > 0 {
		return nil, fmt.Errorf("task %d is not a leaf task", id)
	}

	xp := task.XPValue
	attr := parseStoredAttribute(task.Attribute)

	if err := s.tasks.MarkDone(ctx, id, now); err != nil {
		return nil, err
	}

	p.XPTotal += xp
	addAttributeXP(p, attr, xp)
	p.Level = LevelForTotalXP(p.XPTotal)
	if err := s.players.Update(ctx, p); err != nil {
		return nil, err
	}
	if _, err := s.completions.Insert(ctx, id, now, task.Difficulty, xp); err != nil {
		return nil, err
	}

	levelUp := p.Level > levelBefore
	if levelUp {
		_, _ = s.EvaluateBlueprintUnlocks(ctx)
	}

	return &CompleteResult{
		TaskID:      id,
		XPAwarded:   xp,
		LevelBefore: levelBefore,
		LevelAfter:  p.Level,
		LevelUp:     levelUp,
	}, nil
}

func (s *Service) projectVolumeAndUndone(ctx context.Context, projectID int64) (volume int, hasUndone bool, err error) {
	stack := []int64{projectID}
	seen := map[int64]bool{}

	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[cur] {
			continue
		}
		seen[cur] = true

		children, err := s.tasks.ListChildren(ctx, cur)
		if err != nil {
			return 0, false, err
		}
		for _, c := range children {
			cid := c.ID
			if c.IsProject {
				stack = append(stack, cid)
				continue
			}
			if c.IsHabit {
				// Habits are recurring; they do not block project completion.
				continue
			}
			volume += c.XPValue
			if c.Status != "done" {
				hasUndone = true
			}
		}
	}

	return volume, hasUndone, nil
}

// ProjectHP returns the derived project HP as the sum of xp_value of all unfinished
// (status != done) non-project descendant tasks.
func (s *Service) ProjectHP(ctx context.Context, projectID int64) (int, error) {
	stack := []int64{projectID}
	seen := map[int64]bool{}
	remaining := 0

	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[cur] {
			continue
		}
		seen[cur] = true

		children, err := s.tasks.ListChildren(ctx, cur)
		if err != nil {
			return 0, err
		}
		for _, c := range children {
			cid := c.ID
			if c.IsProject {
				stack = append(stack, cid)
				continue
			}
			if c.IsHabit {
				// Habits have separate recurrence mechanics.
				continue
			}
			if c.Status != "done" {
				remaining += c.XPValue
			}
		}
	}

	return remaining, nil
}
