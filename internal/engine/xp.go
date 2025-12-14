package engine

import (
	"fmt"
	"math"
)

const (
	// XPRequiredCoef is the constant from the spec: XP_req = 500 * (Level^1.5)
	XPRequiredCoef = 500.0

	// TaskBaseXP is the base XP used with difficulty multipliers.
	TaskBaseXP = 50.0

	// AttributeLevelBonusRate is the per-attribute-level bonus rate (5% per level).
	AttributeLevelBonusRate = 0.05
)

// XPRequiredForLevel returns the total XP threshold required to be at the given level.
// Level 0 requires 0 XP.
func XPRequiredForLevel(level int) int {
	if level <= 0 {
		return 0
	}
	req := XPRequiredCoef * math.Pow(float64(level), 1.5)
	// Use ceil to avoid making thresholds easier due to floating point rounding.
	return int(math.Ceil(req))
}

// LevelForTotalXP returns the highest level L such that totalXP >= XPRequiredForLevel(L).
func LevelForTotalXP(totalXP int) int {
	if totalXP <= 0 {
		return 0
	}

	// Exponential search upper bound, then binary search.
	low := 0
	high := 1
	for XPRequiredForLevel(high) <= totalXP {
		low = high
		high *= 2
		if high > 1_000_000 {
			break
		}
	}

	for low+1 < high {
		mid := low + (high-low)/2
		if XPRequiredForLevel(mid) <= totalXP {
			low = mid
		} else {
			high = mid
		}
	}
	return low
}

// AttributeLevelForXP is the same curve as global leveling, applied to attribute-specific XP.
func AttributeLevelForXP(attributeXP int) int {
	return LevelForTotalXP(attributeXP)
}

func difficultyMultiplier(d Difficulty) (float64, error) {
	switch d {
	case DifficultyTrivial:
		return 1.0, nil
	case DifficultyEasy:
		return 2.0, nil
	case DifficultyMedium:
		return 5.0, nil
	case DifficultyHard:
		return 10.0, nil
	case DifficultyEpic:
		return 25.0, nil
	default:
		return 0, fmt.Errorf("invalid difficulty: %d", d)
	}
}

// CalculateXP computes task XP and returns an integer XP value.
// The value is intended to be frozen at task creation time.
func CalculateXP(d Difficulty, attributeLevel int) (int, error) {
	mult, err := difficultyMultiplier(d)
	if err != nil {
		return 0, err
	}
	if attributeLevel < 0 {
		attributeLevel = 0
	}

	base := TaskBaseXP * mult
	bonus := 1.0 + float64(attributeLevel)*AttributeLevelBonusRate
	xp := base * bonus
	// Round to nearest integer for stable results.
	return int(math.Round(xp)), nil
}
