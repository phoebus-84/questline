package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type TaskRepo struct {
	db *sql.DB
}

func NewTaskRepo(db *sql.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

type TaskInsert struct {
	ParentID       *int64
	Title          string
	Description    *string
	Status         string
	DueDate        *time.Time
	Difficulty     int
	Attribute      string
	Attributes     map[string]int // Multi-attribute weights (JSON)
	XPValue        int
	IsProject      bool
	IsHabit        bool
	HabitInterval  *string
	HabitStartDate *time.Time
	HabitEndDate   *time.Time
	HabitGoal      *int
}

func (r *TaskRepo) Insert(ctx context.Context, in TaskInsert) (int64, error) {
	var attrsJSON *string
	if len(in.Attributes) > 0 {
		data, err := json.Marshal(in.Attributes)
		if err != nil {
			return 0, fmt.Errorf("marshal attributes: %w", err)
		}
		s := string(data)
		attrsJSON = &s
	}

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (
			parent_id, title, description,
			status, due_date,
			difficulty, attribute, attributes, xp_value,
			is_project, is_habit, habit_interval,
			habit_start_date, habit_end_date, habit_goal
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, in.ParentID, in.Title, in.Description, in.Status, in.DueDate, in.Difficulty, in.Attribute, attrsJSON, in.XPValue, boolToInt(in.IsProject), boolToInt(in.IsHabit), in.HabitInterval, in.HabitStartDate, in.HabitEndDate, in.HabitGoal)
	if err != nil {
		return 0, fmt.Errorf("task insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("task last insert id: %w", err)
	}
	return id, nil
}

func (r *TaskRepo) Get(ctx context.Context, id int64) (*Task, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, parent_id, title, description, status, created_at, completed_at, due_date,
			difficulty, attribute, attributes, xp_value, is_project, is_habit, habit_interval,
			habit_start_date, habit_end_date, habit_goal
		FROM tasks
		WHERE id = ?
	`, id)

	return scanTaskRow(row)
}

func (r *TaskRepo) ListAll(ctx context.Context) ([]Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, parent_id, title, description, status, created_at, completed_at, due_date,
			difficulty, attribute, attributes, xp_value, is_project, is_habit, habit_interval,
			habit_start_date, habit_end_date, habit_goal
		FROM tasks
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("task list: %w", err)
	}
	defer rows.Close()

	var out []Task
	for rows.Next() {
		t, err := scanTaskRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("task list rows: %w", err)
	}
	return out, nil
}

func (r *TaskRepo) ListChildren(ctx context.Context, parentID int64) ([]Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, parent_id, title, description, status, created_at, completed_at, due_date,
			difficulty, attribute, attributes, xp_value, is_project, is_habit, habit_interval,
			habit_start_date, habit_end_date, habit_goal
		FROM tasks
		WHERE parent_id = ?
		ORDER BY id ASC
	`, parentID)
	if err != nil {
		return nil, fmt.Errorf("task children list: %w", err)
	}
	defer rows.Close()

	var out []Task
	for rows.Next() {
		t, err := scanTaskRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("task children rows: %w", err)
	}
	return out, nil
}

func (r *TaskRepo) MarkDone(ctx context.Context, id int64, completedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE tasks SET status = 'done', completed_at = ? WHERE id = ?`, completedAt, id)
	if err != nil {
		return fmt.Errorf("task mark done: %w", err)
	}
	return nil
}

func (r *TaskRepo) UpdateHabitAfterCompletion(ctx context.Context, id int64, completedAt time.Time, nextDueDate time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE tasks
		SET status = 'pending', completed_at = ?, due_date = ?
		WHERE id = ?
	`, completedAt, nextDueDate, id)
	if err != nil {
		return fmt.Errorf("habit update after completion: %w", err)
	}
	return nil
}

func (r *TaskRepo) UpdateDifficultyAndXP(ctx context.Context, id int64, difficulty int, xpValue int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE tasks SET difficulty = ?, xp_value = ? WHERE id = ?`, difficulty, xpValue, id)
	if err != nil {
		return fmt.Errorf("task update difficulty/xp: %w", err)
	}
	return nil
}

func (r *TaskRepo) HasCompletedProjectTitle(ctx context.Context, title string) (bool, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT 1
		FROM tasks
		WHERE is_project = 1 AND status = 'done' AND title = ?
		LIMIT 1
	`, title)
	var one int
	if err := row.Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("has completed project title: %w", err)
	}
	return true, nil
}

func (r *TaskRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE tasks SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return fmt.Errorf("task update status: %w", err)
	}
	return nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTaskRow(row scanner) (*Task, error) {
	var (
		id             int64
		parent         sql.NullInt64
		title          string
		description    sql.NullString
		status         string
		createdAt      time.Time
		completedAt    sql.NullTime
		dueDate        sql.NullTime
		difficulty     int
		attribute      string
		attributesRaw  sql.NullString
		xpValue        int
		isProject      int
		isHabit        int
		habitInterval  sql.NullString
		habitStartDate sql.NullTime
		habitEndDate   sql.NullTime
		habitGoal      sql.NullInt64
	)

	if err := row.Scan(
		&id, &parent, &title, &description, &status, &createdAt, &completedAt, &dueDate,
		&difficulty, &attribute, &attributesRaw, &xpValue, &isProject, &isHabit, &habitInterval,
		&habitStartDate, &habitEndDate, &habitGoal,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("task scan: %w", err)
	}

	var parentID *int64
	if parent.Valid {
		v := parent.Int64
		parentID = &v
	}
	var desc *string
	if description.Valid {
		v := description.String
		desc = &v
	}
	var comp *time.Time
	if completedAt.Valid {
		v := completedAt.Time
		comp = &v
	}
	var due *time.Time
	if dueDate.Valid {
		v := dueDate.Time
		due = &v
	}
	var interval *string
	if habitInterval.Valid {
		v := habitInterval.String
		interval = &v
	}
	var hStart *time.Time
	if habitStartDate.Valid {
		v := habitStartDate.Time
		hStart = &v
	}
	var hEnd *time.Time
	if habitEndDate.Valid {
		v := habitEndDate.Time
		hEnd = &v
	}
	var hGoal *int
	if habitGoal.Valid {
		v := int(habitGoal.Int64)
		hGoal = &v
	}

	// Parse attributes JSON
	var attrs map[string]int
	if attributesRaw.Valid && attributesRaw.String != "" {
		if err := json.Unmarshal([]byte(attributesRaw.String), &attrs); err != nil {
			return nil, fmt.Errorf("unmarshal attributes: %w", err)
		}
	}

	return &Task{
		ID:             id,
		ParentID:       parentID,
		Title:          title,
		Description:    desc,
		Status:         status,
		CreatedAt:      createdAt,
		CompletedAt:    comp,
		DueDate:        due,
		Difficulty:     difficulty,
		Attribute:      attribute,
		Attributes:     attrs,
		XPValue:        xpValue,
		IsProject:      isProject != 0,
		IsHabit:        isHabit != 0,
		HabitInterval:  interval,
		HabitStartDate: hStart,
		HabitEndDate:   hEnd,
		HabitGoal:      hGoal,
	}, nil
}

func scanTaskRows(rows *sql.Rows) (*Task, error) {
	return scanTaskRow(rows)
}
