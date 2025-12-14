package engine

import (
	"context"
	"fmt"
)

// UpdateTaskDifficulty updates a task/habit difficulty and recalculates xp_value.
// For habits, raising difficulty implicitly resets diminishing returns because the
// decay rule only triggers for repeated completions at the same difficulty.
func (s *Service) UpdateTaskDifficulty(ctx context.Context, id int64, newDifficulty Difficulty) error {
	if !newDifficulty.IsValid() {
		return fmt.Errorf("invalid difficulty: %d", newDifficulty)
	}

	p, err := s.getPlayer(ctx)
	if err != nil {
		return err
	}

	t, err := s.tasks.Get(ctx, id)
	if err != nil {
		return err
	}
	if t == nil {
		return fmt.Errorf("task %d not found", id)
	}
	if t.IsProject {
		return fmt.Errorf("cannot set difficulty on a project")
	}

	attr := parseStoredAttribute(t.Attribute)
	attrXP := playerXPForAttribute(p, attr)
	attrLevel := AttributeLevelForXP(attrXP)

	xpValue, err := CalculateXP(newDifficulty, attrLevel)
	if err != nil {
		return err
	}

	return s.tasks.UpdateDifficultyAndXP(ctx, id, int(newDifficulty), xpValue)
}
