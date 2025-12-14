package engine

import (
	"context"
	"fmt"

	"questline/internal/storage"
)

type CreateTaskInput struct {
	Title      string
	Difficulty Difficulty
	Attribute  Attribute
	ParentID   *int64
	IsHabit    bool
}

type CreateProjectInput struct {
	Title     string
	Attribute Attribute
}

type CreateResult struct {
	TaskID          int64
	ProjectActivated bool
}

type CapacityError struct {
	Limit int
}

func (e CapacityError) Error() string {
	return fmt.Sprintf("too many active tasks (limit %d)", e.Limit)
}

func (s *Service) CreateProject(ctx context.Context, in CreateProjectInput) (*CreateResult, error) {
	title, err := normalizeTitle(in.Title)
	if err != nil {
		return nil, err
	}

	p, err := s.getPlayer(ctx)
	if err != nil {
		return nil, err
	}
	if err := CanCreateProject(p.Level); err != nil {
		return nil, err
	}

	attr := in.Attribute
	if !attr.IsValid() {
		attr = DefaultAttribute
	}

	id, err := s.tasks.Insert(ctx, storage.TaskInsert{
		ParentID:    nil,
		Title:       title,
		Description: nil,
		Status:      "planning",
		DueDate:     nil,
		Difficulty:  int(DifficultyTrivial),
		Attribute:   string(attr),
		XPValue:     0,
		IsProject:   true,
		IsHabit:     false,
		HabitInterval: nil,
	})
	if err != nil {
		return nil, err
	}

	return &CreateResult{TaskID: id}, nil
}

func (s *Service) CreateTask(ctx context.Context, in CreateTaskInput) (*CreateResult, error) {
	title, err := normalizeTitle(in.Title)
	if err != nil {
		return nil, err
	}
	if !in.Difficulty.IsValid() {
		return nil, fmt.Errorf("invalid difficulty: %d", in.Difficulty)
	}

	p, err := s.getPlayer(ctx)
	if err != nil {
		return nil, err
	}

	if in.IsHabit {
		if err := CanCreateHabit(p.Level); err != nil {
			return nil, err
		}
	}

	// Parent validation + gating.
	parentID := in.ParentID
	activated := false
	if parentID != nil {
		parent, err := s.tasks.Get(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, fmt.Errorf("parent task %d not found", *parentID)
		}

		depth, err := s.taskDepthFromRoot(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		// New task will be one deeper than parent.
		if err := CanAttachToParent(p.Level, depth+1); err != nil {
			return nil, err
		}
	}

	activeCount, err := s.countActiveLeafTasks(ctx)
	if err != nil {
		return nil, err
	}
	limit := MaxActiveTasks(p.Level)
	if activeCount >= limit {
		return nil, CapacityError{Limit: limit}
	}

	attr := in.Attribute
	if !attr.IsValid() {
		attr = DefaultAttribute
	}

	attrXP := playerXPForAttribute(p, attr)
	attrLevel := AttributeLevelForXP(attrXP)
	xpValue, err := CalculateXP(in.Difficulty, attrLevel)
	if err != nil {
		return nil, err
	}

	status := "pending"
	if in.IsHabit {
		status = "active"
	}

	id, err := s.tasks.Insert(ctx, storage.TaskInsert{
		ParentID:    parentID,
		Title:       title,
		Description: nil,
		Status:      status,
		DueDate:     nil,
		Difficulty:  int(in.Difficulty),
		Attribute:   string(attr),
		XPValue:     xpValue,
		IsProject:   false,
		IsHabit:     in.IsHabit,
		HabitInterval: nil,
	})
	if err != nil {
		return nil, err
	}

	if parentID != nil {
		parent, err := s.tasks.Get(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if parent != nil && parent.IsProject && parent.Status == "planning" {
			if err := s.tasks.UpdateStatus(ctx, parent.ID, "active"); err != nil {
				return nil, err
			}
			activated = true
		}
	}

	return &CreateResult{TaskID: id, ProjectActivated: activated}, nil
}
