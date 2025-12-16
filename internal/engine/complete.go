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
	TaskID         int64
	XPAwarded      int
	LevelBefore    int
	LevelAfter     int
	LevelUp        bool
	ProjectBonus   bool
	ProjectVolume  int
	HabitCompleted bool // True when a goal-based habit reached its completion target
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

// distributeXP distributes total XP across attributes based on weights.
// If no weights provided (or empty), uses the primary attribute only.
// Weights are percentages (e.g., {STR: 50, INT: 50} means 50% each).
func distributeXP(p *storage.Player, totalXP int, primaryAttr Attribute, weights map[string]int) {
	if len(weights) == 0 {
		// Single attribute mode - all XP goes to primary
		addAttributeXP(p, primaryAttr, totalXP)
		return
	}

	// Calculate total weight
	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}
	if totalWeight == 0 {
		addAttributeXP(p, primaryAttr, totalXP)
		return
	}

	// Distribute XP proportionally
	distributed := 0
	var lastAttr Attribute
	for attrStr, weight := range weights {
		attr := parseStoredAttribute(attrStr)
		share := (totalXP * weight) / totalWeight
		addAttributeXP(p, attr, share)
		distributed += share
		lastAttr = attr
	}

	// Give any remainder (from integer division) to the last attribute
	if remainder := totalXP - distributed; remainder > 0 {
		addAttributeXP(p, lastAttr, remainder)
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
		distributeXP(p, xp, attr, task.Attributes)
		p.Level = LevelForTotalXP(p.XPTotal)
		if err := s.players.Update(ctx, p); err != nil {
			return nil, err
		}
		if _, err := s.completions.Insert(ctx, id, now, task.Difficulty, xp); err != nil {
			return nil, err
		}

		// Check if habit goal is reached
		habitCompleted := false
		if task.HabitGoal != nil {
			allComps, err := s.completions.ListByTask(ctx, id)
			if err != nil {
				return nil, err
			}
			progress := GetHabitProgress(task, allComps, now)
			if progress.Completed {
				// Mark habit as done (goal reached!)
				if err := s.tasks.MarkDone(ctx, id, now); err != nil {
					return nil, err
				}
				habitCompleted = true
			}
		}

		levelUp := p.Level > levelBefore
		if levelUp {
			_, _ = s.EvaluateBlueprintUnlocks(ctx)
		}

		return &CompleteResult{
			TaskID:         id,
			XPAwarded:      xp,
			LevelBefore:    levelBefore,
			LevelAfter:     p.Level,
			LevelUp:        levelUp,
			HabitCompleted: habitCompleted,
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
		distributeXP(p, bonus, attr, task.Attributes)
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
	distributeXP(p, xp, attr, task.Attributes)
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

// RestoreResult holds the result of restoring a task completion.
type RestoreResult struct {
	TaskID      int64
	XPDeducted  int
	LevelBefore int
	LevelAfter  int
	LevelDown   bool
}

// RestoreTask undoes the last completion for a task:
// 1. Finds and deletes the last completion record
// 2. Deducts the XP from the player (total and attribute-specific)
// 3. Resets the task status to "pending"
func (s *Service) RestoreTask(ctx context.Context, id int64) (*RestoreResult, error) {
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

	// Get the last completion for this task
	lastComp, err := s.completions.Last(ctx, id)
	if err != nil {
		return nil, err
	}
	if lastComp == nil {
		return nil, fmt.Errorf("task %d has no completions to restore", id)
	}

	xp := lastComp.XPAwarded

	// Deduct XP from total
	p.XPTotal -= xp
	if p.XPTotal < 0 {
		p.XPTotal = 0
	}

	// Deduct XP from the appropriate attribute
	attr := parseStoredAttribute(task.Attribute)
	if len(task.Attributes) > 0 {
		// If task had multiple attributes, deduct proportionally
		totalWeight := 0
		for _, w := range task.Attributes {
			totalWeight += w
		}
		if totalWeight > 0 {
			for attrStr, weight := range task.Attributes {
				a := parseStoredAttribute(attrStr)
				share := (xp * weight) / totalWeight
				deductAttributeXP(p, a, share)
			}
		}
	} else {
		// Single attribute mode
		deductAttributeXP(p, attr, xp)
	}

	// Recalculate level
	p.Level = LevelForTotalXP(p.XPTotal)

	// Update player
	if err := s.players.Update(ctx, p); err != nil {
		return nil, err
	}

	// Delete the completion record
	if err := s.completions.Delete(ctx, lastComp.ID); err != nil {
		return nil, err
	}

	// Reset task status to pending
	if err := s.tasks.ResetToPending(ctx, id); err != nil {
		return nil, err
	}

	return &RestoreResult{
		TaskID:      id,
		XPDeducted:  xp,
		LevelBefore: levelBefore,
		LevelAfter:  p.Level,
		LevelDown:   p.Level < levelBefore,
	}, nil
}

// deductAttributeXP removes XP from the appropriate attribute on the player.
func deductAttributeXP(p *storage.Player, attr Attribute, xp int) {
	switch attr {
	case AttributeSTR:
		p.XPStr -= xp
		if p.XPStr < 0 {
			p.XPStr = 0
		}
	case AttributeINT:
		p.XPInt -= xp
		if p.XPInt < 0 {
			p.XPInt = 0
		}
	case AttributeART:
		p.XPArt -= xp
		if p.XPArt < 0 {
			p.XPArt = 0
		}
	case AttributeHOME:
		p.XPHome -= xp
		if p.XPHome < 0 {
			p.XPHome = 0
		}
	case AttributeOUT:
		p.XPOut -= xp
		if p.XPOut < 0 {
			p.XPOut = 0
		}
	case AttributeREAD:
		p.XPRead -= xp
		if p.XPRead < 0 {
			p.XPRead = 0
		}
	case AttributeCINEMA:
		p.XPCinema -= xp
		if p.XPCinema < 0 {
			p.XPCinema = 0
		}
	case AttributeCAREER:
		p.XPCareer -= xp
		if p.XPCareer < 0 {
			p.XPCareer = 0
		}
	case AttributeWIS:
		fallthrough
	default:
		p.XPWis -= xp
		if p.XPWis < 0 {
			p.XPWis = 0
		}
	}
}
