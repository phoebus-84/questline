package engine

import "fmt"

const (
	LevelSubtasks   = 3
	LevelHabits     = 5
	LevelProjects   = 7
	LevelDeepRecurs = 10
	LevelReviews    = 15
)

// MaxActiveTasks returns the maximum number of active tasks allowed.
// Spec:
// - Level 0 (Drifter): 3
// - Level 2 (Apprentice): 5
func MaxActiveTasks(level int) int {
	if level >= 2 {
		return 5
	}
	return 3
}

// MaxSubtaskDepth returns the maximum allowed depth for subtasks.
// - If level < 3: no subtasks allowed (depth 0)
// - If 3 <= level < 10: depth 1
// - If level >= 10: unlimited
//
// Depth is measured as: parent-child edge count from the top-level task.
// For gating, the engine will compare the would-be depth of the new task.
const SubtaskDepthUnlimited = -1

func MaxSubtaskDepth(level int) int {
	if level < LevelSubtasks {
		return 0
	}
	if level < LevelDeepRecurs {
		return 1
	}
	return SubtaskDepthUnlimited
}

func CanCreateHabit(level int) error {
	if level < LevelHabits {
		return GateError{Feature: "habits", RequiredLevel: LevelHabits}
	}
	return nil
}

func CanCreateProject(level int) error {
	if level < LevelProjects {
		return GateError{Feature: "projects", RequiredLevel: LevelProjects}
	}
	return nil
}

func CanAttachToParent(level int, requestedDepth int) error {
	maxDepth := MaxSubtaskDepth(level)
	if maxDepth == SubtaskDepthUnlimited {
		return nil
	}
	if requestedDepth <= 0 {
		return nil
	}
	if maxDepth == 0 {
		return GateError{Feature: "subtasks", RequiredLevel: LevelSubtasks}
	}
	if requestedDepth > maxDepth {
		return fmt.Errorf("subtask depth %d exceeds max depth %d at level %d", requestedDepth, maxDepth, level)
	}
	return nil
}
