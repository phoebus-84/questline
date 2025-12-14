package engine

import (
	"context"
	"path/filepath"
	"testing"

	"questline/internal/storage"
)

func newTestService(t *testing.T) (*Service, func()) {
	t.Helper()
	ctx := context.Background()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	db, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	svc := NewService(db)
	cleanup := func() {
		_ = db.Close()
	}
	return svc, cleanup
}

func setPlayerXP(t *testing.T, svc *Service, totalXP int) {
	t.Helper()
	ctx := context.Background()
	p, err := svc.PlayerRepo().GetOrCreateMain(ctx)
	if err != nil {
		t.Fatalf("get player: %v", err)
	}
	p.XPTotal = totalXP
	if err := svc.PlayerRepo().Update(ctx, p); err != nil {
		t.Fatalf("update player: %v", err)
	}
}

func TestXPBoundaries(t *testing.T) {
	if got := XPRequiredForLevel(0); got != 0 {
		t.Fatalf("XPRequiredForLevel(0)=%d, want 0", got)
	}
	l1 := XPRequiredForLevel(1)
	if got := LevelForTotalXP(l1 - 1); got != 0 {
		t.Fatalf("LevelForTotalXP(l1-1)=%d, want 0", got)
	}
	if got := LevelForTotalXP(l1); got != 1 {
		t.Fatalf("LevelForTotalXP(l1)=%d, want 1", got)
	}

	l7 := XPRequiredForLevel(7)
	if got := LevelForTotalXP(l7); got != 7 {
		t.Fatalf("LevelForTotalXP(l7)=%d, want 7", got)
	}
}

func TestProjectPlanningActivationAndCompletionBonus(t *testing.T) {
	svc, cleanup := newTestService(t)
	defer cleanup()
	ctx := context.Background()

	setPlayerXP(t, svc, XPRequiredForLevel(LevelProjects))

	proj, err := svc.CreateProject(ctx, CreateProjectInput{Title: "Read a Book", Attribute: AttributeART})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if _, err := svc.CompleteTask(ctx, proj.TaskID); err == nil {
		t.Fatalf("expected error completing planning project")
	}

	pid := proj.TaskID
	child1, err := svc.CreateTask(ctx, CreateTaskInput{
		Title:      "Pick a book",
		Difficulty: DifficultyEasy,
		Attribute:  AttributeWIS,
		ParentID:   &pid,
		IsHabit:    false,
	})
	if err != nil {
		t.Fatalf("CreateTask child1: %v", err)
	}
	if !child1.ProjectActivated {
		t.Fatalf("expected project activation on first child")
	}

	child2, err := svc.CreateTask(ctx, CreateTaskInput{
		Title:      "Read 30 pages",
		Difficulty: DifficultyMedium,
		Attribute:  AttributeART,
		ParentID:   &pid,
		IsHabit:    false,
	})
	if err != nil {
		t.Fatalf("CreateTask child2: %v", err)
	}
	if child2.ProjectActivated {
		t.Fatalf("did not expect project activation on second child")
	}

	pTask, err := svc.TaskRepo().Get(ctx, pid)
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	if pTask.Status != "active" {
		t.Fatalf("project status=%q, want active", pTask.Status)
	}

	if _, err := svc.CompleteTask(ctx, child1.TaskID); err != nil {
		t.Fatalf("complete child1: %v", err)
	}
	if _, err := svc.CompleteTask(ctx, child2.TaskID); err != nil {
		t.Fatalf("complete child2: %v", err)
	}

	c1, _ := svc.TaskRepo().Get(ctx, child1.TaskID)
	c2, _ := svc.TaskRepo().Get(ctx, child2.TaskID)
	volume := c1.XPValue + c2.XPValue
	wantBonus := int(float64(volume)*0.10 + 0.5)

	res, err := svc.CompleteTask(ctx, pid)
	if err != nil {
		t.Fatalf("complete project: %v", err)
	}
	if !res.ProjectBonus {
		t.Fatalf("expected ProjectBonus=true")
	}
	if res.XPAwarded != wantBonus {
		t.Fatalf("bonus xp=%d, want %d (volume=%d)", res.XPAwarded, wantBonus, volume)
	}
}

func TestHabitDecayAndDifficultyReset(t *testing.T) {
	svc, cleanup := newTestService(t)
	defer cleanup()
	ctx := context.Background()

	setPlayerXP(t, svc, XPRequiredForLevel(LevelHabits))

	h, err := svc.CreateTask(ctx, CreateTaskInput{
		Title:         "Push-ups",
		Difficulty:    DifficultyEasy,
		Attribute:     AttributeSTR,
		IsHabit:       true,
		HabitInterval: HabitIntervalDaily,
	})
	if err != nil {
		t.Fatalf("CreateHabit: %v", err)
	}

	habit, err := svc.TaskRepo().Get(ctx, h.TaskID)
	if err != nil {
		t.Fatalf("get habit: %v", err)
	}
	base := habit.XPValue

	for i := 0; i < 5; i++ {
		res, err := svc.CompleteTask(ctx, h.TaskID)
		if err != nil {
			t.Fatalf("complete habit #%d: %v", i+1, err)
		}
		if res.XPAwarded != base {
			t.Fatalf("habit xp #%d=%d, want %d", i+1, res.XPAwarded, base)
		}
	}

	res6, err := svc.CompleteTask(ctx, h.TaskID)
	if err != nil {
		t.Fatalf("complete habit #6: %v", err)
	}
	wantHalf := int(float64(base)*0.50 + 0.5)
	if wantHalf < 1 {
		wantHalf = 1
	}
	if res6.XPAwarded != wantHalf {
		t.Fatalf("habit xp #6=%d, want %d", res6.XPAwarded, wantHalf)
	}

	if err := svc.UpdateTaskDifficulty(ctx, h.TaskID, DifficultyMedium); err != nil {
		t.Fatalf("UpdateTaskDifficulty: %v", err)
	}
	updated, err := svc.TaskRepo().Get(ctx, h.TaskID)
	if err != nil {
		t.Fatalf("get updated habit: %v", err)
	}
	base2 := updated.XPValue

	resAfter, err := svc.CompleteTask(ctx, h.TaskID)
	if err != nil {
		t.Fatalf("complete habit after diff increase: %v", err)
	}
	if resAfter.XPAwarded != base2 {
		t.Fatalf("habit xp after diff increase=%d, want %d", resAfter.XPAwarded, base2)
	}

	final, err := svc.TaskRepo().Get(ctx, h.TaskID)
	if err != nil {
		t.Fatalf("get habit final: %v", err)
	}
	if final.Status != "pending" {
		t.Fatalf("habit status=%q, want pending", final.Status)
	}
	if final.DueDate == nil {
		t.Fatalf("expected habit due_date to be set")
	}
}
