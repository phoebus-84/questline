package engine

import "fmt"

const (
	LevelSubtasks   = 3
	LevelHabits     = 5
	LevelProjects   = 7
	LevelDeepRecurs = 10
	LevelReviews    = 15
)

// DifficultyUnlockLevels maps difficulty levels to the player level required.
// Players start with difficulty 1 only. Higher difficulties unlock as they level up.
var DifficultyUnlockLevels = map[Difficulty]int{
	DifficultyTrivial: 0,  // Always available
	DifficultyEasy:    2,  // Level 2+
	DifficultyMedium:  5,  // Level 5+
	DifficultyHard:    8,  // Level 8+
	DifficultyEpic:    12, // Level 12+
}

// MaxDifficultyForLevel returns the highest difficulty available at the given player level.
func MaxDifficultyForLevel(level int) Difficulty {
	max := DifficultyTrivial
	for diff, req := range DifficultyUnlockLevels {
		if level >= req && diff > max {
			max = diff
		}
	}
	return max
}

// CanUseDifficulty returns an error if the player level is too low for the requested difficulty.
func CanUseDifficulty(level int, difficulty Difficulty) error {
	reqLevel, ok := DifficultyUnlockLevels[difficulty]
	if !ok {
		return fmt.Errorf("invalid difficulty: %d", difficulty)
	}
	if level < reqLevel {
		return DifficultyGateError{
			Difficulty:    difficulty,
			RequiredLevel: reqLevel,
			CurrentLevel:  level,
		}
	}
	return nil
}

// DifficultyGateError is returned when a player tries to use a locked difficulty.
type DifficultyGateError struct {
	Difficulty    Difficulty
	RequiredLevel int
	CurrentLevel  int
}

func (e DifficultyGateError) Error() string {
	return fmt.Sprintf("difficulty %d requires level %d (currently %d)", e.Difficulty, e.RequiredLevel, e.CurrentLevel)
}

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
