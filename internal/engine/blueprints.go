package engine

import (
	"context"
	"fmt"
	"strings"

	"questline/internal/storage"
)

type BlueprintStatus string

const (
	BlueprintLocked    BlueprintStatus = "locked"
	BlueprintAvailable BlueprintStatus = "available"
	BlueprintActive    BlueprintStatus = "active"
	BlueprintCompleted BlueprintStatus = "completed"
)

type BlueprintKind string

const (
	BlueprintKindTask    BlueprintKind = "task"
	BlueprintKindProject BlueprintKind = "project"
	BlueprintKindHabit   BlueprintKind = "habit"
)

// PlayerAttrLevel returns the attribute level for a given attribute from a player.
func PlayerAttrLevel(p *storage.Player, attr Attribute) int {
	switch attr {
	case AttributeSTR:
		return AttributeLevelForXP(p.XPStr)
	case AttributeINT:
		return AttributeLevelForXP(p.XPInt)
	case AttributeWIS:
		return AttributeLevelForXP(p.XPWis)
	case AttributeART:
		return AttributeLevelForXP(p.XPArt)
	case AttributeHOME:
		return AttributeLevelForXP(p.XPHome)
	case AttributeOUT:
		return AttributeLevelForXP(p.XPOut)
	case AttributeREAD:
		return AttributeLevelForXP(p.XPRead)
	case AttributeCINEMA:
		return AttributeLevelForXP(p.XPCinema)
	case AttributeCAREER:
		return AttributeLevelForXP(p.XPCareer)
	default:
		return 0
	}
}

// UnlockReq represents a single attribute level requirement.
type UnlockReq struct {
	Attr     Attribute
	MinLevel int
}

// CheckUnlockReqs checks if all attribute requirements are met for a player.
func CheckUnlockReqs(p *storage.Player, reqs ...UnlockReq) bool {
	for _, req := range reqs {
		if PlayerAttrLevel(p, req.Attr) < req.MinLevel {
			return false
		}
	}
	return true
}

// BlueprintChild represents a child task/habit to create when a project blueprint is accepted.
type BlueprintChild struct {
	Title      string
	Difficulty Difficulty
	Attribute  Attribute
	IsHabit    bool
	HabitEvery HabitInterval
}

type BlueprintDef struct {
	Code        string
	Kind        BlueprintKind
	Description string

	Title      string
	Difficulty Difficulty
	Attribute  Attribute
	HabitEvery HabitInterval
	Children   []BlueprintChild // Children to auto-create on accept (for projects)

	Unlock func(ctx context.Context, svc *Service, p *storage.Player) (bool, error)
}

func builtinBlueprints() []BlueprintDef {
	return []BlueprintDef{
		// ========== STR (Strength/Fitness) ==========
		{
			Code:        "str_starter",
			Kind:        BlueprintKindHabit,
			Description: "Build a daily push-up habit to strengthen your body and willpower.",
			Title:       "Push-ups",
			Difficulty:  DifficultyEasy,
			Attribute:   AttributeSTR,
			HabitEvery:  HabitIntervalDaily,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelHabits, nil
			},
		},
		{
			Code:        "str_walk",
			Kind:        BlueprintKindHabit,
			Description: "Take a daily walk to clear your mind and stay active.",
			Title:       "Daily Walk",
			Difficulty:  DifficultyTrivial,
			Attribute:   AttributeSTR,
			HabitEvery:  HabitIntervalDaily,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelHabits, nil
			},
		},
		{
			Code:        "str_run",
			Kind:        BlueprintKindHabit,
			Description: "Weekly running builds endurance and cardiovascular health.",
			Title:       "Weekly Run",
			Difficulty:  DifficultyMedium,
			Attribute:   AttributeSTR,
			HabitEvery:  HabitIntervalWeekly,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= 5 && PlayerAttrLevel(p, AttributeSTR) >= 2, nil
			},
		},
		{
			Code:        "str_gym",
			Kind:        BlueprintKindProject,
			Description: "Complete a structured 4-week gym program with progressive overload.",
			Title:       "Gym Program",
			Attribute:   AttributeSTR,
			Children: []BlueprintChild{
				{Title: "Week 1: Foundation", Difficulty: DifficultyEasy},
				{Title: "Week 2: Building", Difficulty: DifficultyMedium},
				{Title: "Week 3: Intensity", Difficulty: DifficultyMedium},
				{Title: "Week 4: Peak", Difficulty: DifficultyHard},
			},
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= 8 && PlayerAttrLevel(p, AttributeSTR) >= 3, nil
			},
		},

		// ========== INT (Intelligence/Learning) ==========
		{
			Code:        "int_puzzle",
			Kind:        BlueprintKindHabit,
			Description: "Solve a puzzle or brain teaser daily to keep your mind sharp.",
			Title:       "Daily Puzzle",
			Difficulty:  DifficultyEasy,
			Attribute:   AttributeINT,
			HabitEvery:  HabitIntervalDaily,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelHabits, nil
			},
		},
		{
			Code:        "int_course",
			Kind:        BlueprintKindProject,
			Description: "Complete an online course on a topic that interests you.",
			Title:       "Online Course",
			Attribute:   AttributeINT,
			Children: []BlueprintChild{
				{Title: "Module 1", Difficulty: DifficultyEasy},
				{Title: "Module 2", Difficulty: DifficultyMedium},
				{Title: "Module 3", Difficulty: DifficultyMedium},
				{Title: "Final Project", Difficulty: DifficultyHard},
			},
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelProjects && PlayerAttrLevel(p, AttributeINT) >= 2, nil
			},
		},
		{
			Code:        "int_lang",
			Kind:        BlueprintKindHabit,
			Description: "Practice a new language daily with apps or study sessions.",
			Title:       "Language Practice",
			Difficulty:  DifficultyMedium,
			Attribute:   AttributeINT,
			HabitEvery:  HabitIntervalDaily,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= 6 && PlayerAttrLevel(p, AttributeINT) >= 2, nil
			},
		},

		// ========== WIS (Wisdom/Mindfulness) ==========
		{
			Code:        "wis_meditate",
			Kind:        BlueprintKindHabit,
			Description: "Daily meditation to cultivate presence and inner calm.",
			Title:       "Meditation",
			Difficulty:  DifficultyEasy,
			Attribute:   AttributeWIS,
			HabitEvery:  HabitIntervalDaily,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelHabits, nil
			},
		},
		{
			Code:        "wis_journal",
			Kind:        BlueprintKindHabit,
			Description: "Write in a journal daily to process thoughts and gain clarity.",
			Title:       "Daily Journal",
			Difficulty:  DifficultyEasy,
			Attribute:   AttributeWIS,
			HabitEvery:  HabitIntervalDaily,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelHabits && PlayerAttrLevel(p, AttributeWIS) >= 1, nil
			},
		},
		{
			Code:        "wis_digital_detox",
			Kind:        BlueprintKindTask,
			Description: "Spend a full day without screens to reset your attention.",
			Title:       "Digital Detox Day",
			Difficulty:  DifficultyHard,
			Attribute:   AttributeWIS,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= 8 && PlayerAttrLevel(p, AttributeWIS) >= 3, nil
			},
		},

		// ========== ART (Creativity) ==========
		{
			Code:        "art_reader",
			Kind:        BlueprintKindProject,
			Description: "Choose a book and read it cover to cover. Track chapters as subtasks.",
			Title:       "Read a Book",
			Attribute:   AttributeART,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				artLevel := AttributeLevelForXP(p.XPArt)
				return p.Level >= LevelProjects && artLevel >= 1, nil
			},
		},
		{
			Code:        "art_critic",
			Kind:        BlueprintKindTask,
			Description: "Reflect on your reading and write a short review to solidify your thoughts.",
			Title:       "Write a short review",
			Difficulty:  DifficultyMedium,
			Attribute:   AttributeART,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				artLevel := AttributeLevelForXP(p.XPArt)
				if p.Level < LevelProjects || artLevel < 2 {
					return false, nil
				}
				has, err := svc.tasks.HasCompletedProjectTitle(ctx, "Read a Book")
				if err != nil {
					return false, err
				}
				return has, nil
			},
		},
		{
			Code:        "art_sketch",
			Kind:        BlueprintKindHabit,
			Description: "Sketch something daily to develop your visual creativity.",
			Title:       "Daily Sketch",
			Difficulty:  DifficultyEasy,
			Attribute:   AttributeART,
			HabitEvery:  HabitIntervalDaily,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelHabits && PlayerAttrLevel(p, AttributeART) >= 1, nil
			},
		},
		{
			Code:        "art_music",
			Kind:        BlueprintKindHabit,
			Description: "Practice an instrument weekly to develop musical skills.",
			Title:       "Music Practice",
			Difficulty:  DifficultyMedium,
			Attribute:   AttributeART,
			HabitEvery:  HabitIntervalWeekly,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= 5 && PlayerAttrLevel(p, AttributeART) >= 2, nil
			},
		},

		// ========== HOME (Household) ==========
		{
			Code:        "home_tidy",
			Kind:        BlueprintKindHabit,
			Description: "Spend 10 minutes tidying your space daily for a calmer environment.",
			Title:       "Daily Tidy",
			Difficulty:  DifficultyTrivial,
			Attribute:   AttributeHOME,
			HabitEvery:  HabitIntervalDaily,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelHabits, nil
			},
		},
		{
			Code:        "home_declutter",
			Kind:        BlueprintKindProject,
			Description: "Go through each room and declutter systematically.",
			Title:       "Home Declutter",
			Attribute:   AttributeHOME,
			Children: []BlueprintChild{
				{Title: "Kitchen", Difficulty: DifficultyMedium},
				{Title: "Bedroom", Difficulty: DifficultyMedium},
				{Title: "Living Room", Difficulty: DifficultyMedium},
				{Title: "Storage Areas", Difficulty: DifficultyHard},
			},
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelProjects && PlayerAttrLevel(p, AttributeHOME) >= 2, nil
			},
		},
		{
			Code:        "home_cook",
			Kind:        BlueprintKindHabit,
			Description: "Cook a homemade meal weekly to improve health and save money.",
			Title:       "Weekly Cooking",
			Difficulty:  DifficultyMedium,
			Attribute:   AttributeHOME,
			HabitEvery:  HabitIntervalWeekly,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelHabits && PlayerAttrLevel(p, AttributeHOME) >= 1, nil
			},
		},

		// ========== OUT (Outdoors/Social) ==========
		{
			Code:        "out_nature",
			Kind:        BlueprintKindHabit,
			Description: "Spend time in nature weekly to recharge and connect with the world.",
			Title:       "Nature Time",
			Difficulty:  DifficultyEasy,
			Attribute:   AttributeOUT,
			HabitEvery:  HabitIntervalWeekly,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelHabits, nil
			},
		},
		{
			Code:        "out_social",
			Kind:        BlueprintKindHabit,
			Description: "Meet with friends weekly to maintain meaningful relationships.",
			Title:       "Friend Meetup",
			Difficulty:  DifficultyEasy,
			Attribute:   AttributeOUT,
			HabitEvery:  HabitIntervalWeekly,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelHabits && PlayerAttrLevel(p, AttributeOUT) >= 1, nil
			},
		},
		{
			Code:        "out_explore",
			Kind:        BlueprintKindTask,
			Description: "Visit a new place in your city you've never been to before.",
			Title:       "Explore New Place",
			Difficulty:  DifficultyMedium,
			Attribute:   AttributeOUT,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= 4 && PlayerAttrLevel(p, AttributeOUT) >= 1, nil
			},
		},

		// ========== READ (Reading) ==========
		{
			Code:        "read_chapter",
			Kind:        BlueprintKindHabit,
			Description: "Read at least one chapter of a book daily.",
			Title:       "Daily Reading",
			Difficulty:  DifficultyEasy,
			Attribute:   AttributeREAD,
			HabitEvery:  HabitIntervalDaily,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelHabits, nil
			},
		},
		{
			Code:        "read_classic",
			Kind:        BlueprintKindProject,
			Description: "Read a literary classic that has stood the test of time.",
			Title:       "Read a Classic",
			Attribute:   AttributeREAD,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelProjects && PlayerAttrLevel(p, AttributeREAD) >= 2, nil
			},
		},
		{
			Code:        "read_nonfiction",
			Kind:        BlueprintKindProject,
			Description: "Read a non-fiction book to expand your knowledge.",
			Title:       "Non-Fiction Deep Dive",
			Attribute:   AttributeREAD,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelProjects && PlayerAttrLevel(p, AttributeREAD) >= 1, nil
			},
		},

		// ========== CINEMA (Film/Culture) ==========
		{
			Code:        "cinema_weekly",
			Kind:        BlueprintKindHabit,
			Description: "Watch a film weekly and reflect on its themes and craft.",
			Title:       "Weekly Film",
			Difficulty:  DifficultyEasy,
			Attribute:   AttributeCINEMA,
			HabitEvery:  HabitIntervalWeekly,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelHabits, nil
			},
		},
		{
			Code:        "cinema_director",
			Kind:        BlueprintKindProject,
			Description: "Watch the filmography of a famous director to understand their vision.",
			Title:       "Director Study",
			Attribute:   AttributeCINEMA,
			Children: []BlueprintChild{
				{Title: "Early Work", Difficulty: DifficultyEasy},
				{Title: "Breakthrough Films", Difficulty: DifficultyMedium},
				{Title: "Masterpieces", Difficulty: DifficultyMedium},
				{Title: "Recent Work", Difficulty: DifficultyEasy},
			},
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelProjects && PlayerAttrLevel(p, AttributeCINEMA) >= 2, nil
			},
		},
		{
			Code:        "cinema_theater",
			Kind:        BlueprintKindTask,
			Description: "Experience a film on the big screen at a theater.",
			Title:       "Theater Visit",
			Difficulty:  DifficultyMedium,
			Attribute:   AttributeCINEMA,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= 3 && PlayerAttrLevel(p, AttributeCINEMA) >= 1, nil
			},
		},

		// ========== CAREER (Professional) ==========
		{
			Code:        "career_network",
			Kind:        BlueprintKindHabit,
			Description: "Reach out to one professional contact monthly to maintain your network.",
			Title:       "Monthly Networking",
			Difficulty:  DifficultyMedium,
			Attribute:   AttributeCAREER,
			HabitEvery:  HabitIntervalMonthly,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelHabits && PlayerAttrLevel(p, AttributeCAREER) >= 1, nil
			},
		},
		{
			Code:        "career_skill",
			Kind:        BlueprintKindProject,
			Description: "Learn a new professional skill through deliberate practice.",
			Title:       "New Skill",
			Attribute:   AttributeCAREER,
			Children: []BlueprintChild{
				{Title: "Research & Plan", Difficulty: DifficultyEasy},
				{Title: "Basic Practice", Difficulty: DifficultyMedium},
				{Title: "Intermediate Practice", Difficulty: DifficultyMedium},
				{Title: "Apply in Real Project", Difficulty: DifficultyHard},
			},
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= LevelProjects && PlayerAttrLevel(p, AttributeCAREER) >= 2, nil
			},
		},
		{
			Code:        "career_resume",
			Kind:        BlueprintKindTask,
			Description: "Update your resume with recent accomplishments and skills.",
			Title:       "Update Resume",
			Difficulty:  DifficultyMedium,
			Attribute:   AttributeCAREER,
			Unlock: func(ctx context.Context, svc *Service, p *storage.Player) (bool, error) {
				return p.Level >= 4, nil
			},
		},
	}
}

func normalizeBlueprintCode(code string) (string, error) {
	c := strings.TrimSpace(strings.ToLower(code))
	if c == "" {
		return "", fmt.Errorf("blueprint code is required")
	}
	return c, nil
}

// GetBlueprintDef returns the blueprint definition by code, or nil if not found.
func GetBlueprintDef(code string) *BlueprintDef {
	defs := builtinBlueprints()
	for i := range defs {
		if defs[i].Code == code {
			return &defs[i]
		}
	}
	return nil
}

// EvaluateBlueprintUnlocks ensures built-in blueprint rows exist and transitions any
// newly unlocked ones from locked -> available.
func (s *Service) EvaluateBlueprintUnlocks(ctx context.Context) ([]storage.Blueprint, error) {
	p, err := s.getPlayer(ctx)
	if err != nil {
		return nil, err
	}

	defs := builtinBlueprints()
	var newlyAvailable []storage.Blueprint

	for _, def := range defs {
		row, err := s.blueprints.Get(ctx, def.Code)
		if err != nil {
			return nil, err
		}
		if row == nil {
			if err := s.blueprints.Upsert(ctx, storage.Blueprint{Code: def.Code, Status: string(BlueprintLocked)}); err != nil {
				return nil, err
			}
			row = &storage.Blueprint{Code: def.Code, Status: string(BlueprintLocked)}
		}
		if row.Status != string(BlueprintLocked) {
			continue
		}

		ok, err := def.Unlock(ctx, s, p)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		b := storage.Blueprint{Code: def.Code, Status: string(BlueprintAvailable)}
		if err := s.blueprints.Upsert(ctx, b); err != nil {
			return nil, err
		}
		newlyAvailable = append(newlyAvailable, b)
	}

	return newlyAvailable, nil
}

// AcceptBlueprint marks an available blueprint as active and instantiates its
// corresponding task/project/habit.
func (s *Service) AcceptBlueprint(ctx context.Context, code string) (*CreateResult, error) {
	c, err := normalizeBlueprintCode(code)
	if err != nil {
		return nil, err
	}

	// Make sure unlock state is up to date (important if player imported/migrated state).
	if _, err := s.EvaluateBlueprintUnlocks(ctx); err != nil {
		return nil, err
	}

	b, err := s.blueprints.Get(ctx, c)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, fmt.Errorf("unknown blueprint: %s", c)
	}
	if b.Status != string(BlueprintAvailable) {
		return nil, fmt.Errorf("blueprint %s is not available (status=%s)", c, b.Status)
	}

	defs := builtinBlueprints()
	var def *BlueprintDef
	for i := range defs {
		if defs[i].Code == c {
			def = &defs[i]
			break
		}
	}
	if def == nil {
		return nil, fmt.Errorf("unknown blueprint: %s", c)
	}

	var res *CreateResult
	switch def.Kind {
	case BlueprintKindProject:
		res, err = s.CreateProject(ctx, CreateProjectInput{Title: def.Title, Attribute: def.Attribute})
	case BlueprintKindHabit:
		res, err = s.CreateTask(ctx, CreateTaskInput{Title: def.Title, Difficulty: def.Difficulty, Attribute: def.Attribute, IsHabit: true, HabitInterval: def.HabitEvery})
	case BlueprintKindTask:
		res, err = s.CreateTask(ctx, CreateTaskInput{Title: def.Title, Difficulty: def.Difficulty, Attribute: def.Attribute, IsHabit: false})
	default:
		return nil, fmt.Errorf("invalid blueprint kind: %s", def.Kind)
	}
	if err != nil {
		return nil, err
	}

	// Auto-create children (for project blueprints)
	if def.Kind == BlueprintKindProject && len(def.Children) > 0 {
		for _, child := range def.Children {
			attr := child.Attribute
			if attr == "" {
				attr = def.Attribute // inherit from parent
			}
			_, err := s.CreateTask(ctx, CreateTaskInput{
				Title:         child.Title,
				Difficulty:    child.Difficulty,
				Attribute:     attr,
				IsHabit:       child.IsHabit,
				HabitInterval: child.HabitEvery,
				ParentID:      &res.TaskID,
			})
			if err != nil {
				return nil, fmt.Errorf("creating child %q: %w", child.Title, err)
			}
		}
	}

	if err := s.blueprints.Upsert(ctx, storage.Blueprint{Code: c, Status: string(BlueprintActive)}); err != nil {
		return nil, err
	}

	return res, nil
}
