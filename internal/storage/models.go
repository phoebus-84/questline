package storage

import "time"

type Player struct {
	Key     string
	Level   int
	XPTotal int
	XPStr   int
	XPInt   int
	XPWis   int
	XPArt   int
}

type Task struct {
	ID            int64
	ParentID      *int64
	Title         string
	Description   *string
	Status        string
	CreatedAt     time.Time
	CompletedAt   *time.Time
	DueDate       *time.Time
	Difficulty    int
	Attribute     string
	XPValue       int
	IsProject     bool
	IsHabit       bool
	HabitInterval *string
}

type Blueprint struct {
	Code   string
	Status string
}

type TaskCompletion struct {
	ID          int64
	TaskID      int64
	CompletedAt time.Time
	Difficulty  int
	XPAwarded   int
}
