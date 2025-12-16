package engine

import (
	"fmt"
	"strings"
	"time"
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
