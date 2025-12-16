package engine

import (
	"fmt"
	"strings"
	"time"

	"questline/internal/storage"
)

type HabitInterval string

const (
	HabitIntervalDaily   HabitInterval = "daily"
	HabitIntervalWeekly  HabitInterval = "weekly"
	HabitIntervalMonthly HabitInterval = "monthly"
)

func (h HabitInterval) IsValid() bool {
	switch h {
	case HabitIntervalDaily, HabitIntervalWeekly, HabitIntervalMonthly:
		return true
	default:
		return false
	}
}

func ParseHabitInterval(input string) (HabitInterval, error) {
	s := strings.TrimSpace(strings.ToLower(input))
	h := HabitInterval(s)
	if !h.IsValid() {
		return "", fmt.Errorf("invalid habit interval: %q", input)
	}
	return h, nil
}

func NextDueDate(now time.Time, interval HabitInterval) (time.Time, error) {
	switch interval {
	case HabitIntervalDaily:
		return now.Add(24 * time.Hour), nil
	case HabitIntervalWeekly:
		return now.Add(7 * 24 * time.Hour), nil
	case HabitIntervalMonthly:
		return now.AddDate(0, 1, 0), nil
	default:
		return time.Time{}, fmt.Errorf("invalid habit interval: %q", interval)
	}
}

// HabitProgress represents the current progress of a timed/goal habit.
type HabitProgress struct {
	Completions int        // Number of completions so far
	Goal        *int       // Target completions (nil = ongoing)
	StartDate   *time.Time // When habit started
	EndDate     *time.Time // When habit ends (nil = forever)
	Completed   bool       // Whether habit has reached its goal
	Expired     bool       // Whether habit has expired without reaching goal
}

// GetHabitProgress calculates the current progress of a habit.
func GetHabitProgress(task *storage.Task, completions []storage.TaskCompletion, now time.Time) HabitProgress {
	progress := HabitProgress{
		Goal:      task.HabitGoal,
		StartDate: task.HabitStartDate,
		EndDate:   task.HabitEndDate,
	}

	// Count completions within the duration window
	startDate := task.HabitStartDate
	if startDate == nil {
		// If no start date, use task creation date
		startDate = &task.CreatedAt
	}

	for _, c := range completions {
		// Only count completions within the duration window
		if c.CompletedAt.Before(*startDate) {
			continue
		}
		if task.HabitEndDate != nil && c.CompletedAt.After(*task.HabitEndDate) {
			continue
		}
		progress.Completions++
	}

	// Check if goal is reached
	if task.HabitGoal != nil && progress.Completions >= *task.HabitGoal {
		progress.Completed = true
	}

	// Check if expired (end date passed without reaching goal)
	if task.HabitEndDate != nil && now.After(*task.HabitEndDate) && !progress.Completed {
		progress.Expired = true
	}

	return progress}