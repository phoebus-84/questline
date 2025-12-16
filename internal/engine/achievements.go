package engine

import (
	"context"

	"questline/internal/storage"
)

// Achievement represents a badge/achievement the player can earn.
type Achievement struct {
	ID          string
	Name        string
	Description string
	Icon        string
	Earned      bool
}

// AchievementChecker calculates which achievements the player has earned.
type AchievementChecker struct {
	player     *storage.Player
	tasks      []storage.Task
	blueprints []storage.Blueprint
}

func NewAchievementChecker(player *storage.Player, tasks []storage.Task, blueprints []storage.Blueprint) *AchievementChecker {
	return &AchievementChecker{
		player:     player,
		tasks:      tasks,
		blueprints: blueprints,
	}
}

// GetAchievements returns all achievements with their earned status.
func (c *AchievementChecker) GetAchievements() []Achievement {
	achievements := []Achievement{
		// Level milestones
		c.levelAchievement("first_steps", "First Steps", "Reach level 1", "🌱", 1),
		c.levelAchievement("getting_started", "Getting Started", "Reach level 3", "🌿", 3),
		c.levelAchievement("on_the_path", "On the Path", "Reach level 5", "🌳", 5),
		c.levelAchievement("seasoned", "Seasoned Adventurer", "Reach level 10", "⭐", 10),
		c.levelAchievement("veteran", "Veteran", "Reach level 15", "🌟", 15),
		c.levelAchievement("master", "Master", "Reach level 20", "💫", 20),

		// Task completion milestones
		c.taskCountAchievement("first_task", "First Quest", "Complete 1 task", "✓", 1),
		c.taskCountAchievement("productive", "Productive", "Complete 10 tasks", "📋", 10),
		c.taskCountAchievement("achiever", "Achiever", "Complete 50 tasks", "🏅", 50),
		c.taskCountAchievement("powerhouse", "Powerhouse", "Complete 100 tasks", "🏆", 100),

		// Attribute level achievements
		c.attrLevelAchievement("strong", "Strong", "STR level 3", "💪", "str", 3),
		c.attrLevelAchievement("smart", "Smart", "INT level 3", "🧠", "int", 3),
		c.attrLevelAchievement("wise", "Wise", "WIS level 3", "🧘", "wis", 3),
		c.attrLevelAchievement("creative", "Creative", "ART level 3", "🎨", "art", 3),
		c.attrLevelAchievement("homemaker", "Homemaker", "HOME level 3", "🏠", "home", 3),
		c.attrLevelAchievement("outdoorsy", "Outdoorsy", "OUT level 3", "🌲", "out", 3),
		c.attrLevelAchievement("bookworm", "Bookworm", "READ level 3", "📚", "read", 3),
		c.attrLevelAchievement("cinephile", "Cinephile", "CINEMA level 3", "🎬", "cinema", 3),
		c.attrLevelAchievement("professional", "Professional", "CAREER level 3", "💼", "career", 3),

		// Blueprint completions
		c.blueprintAchievement("first_blueprint", "Quest Accepted", "Complete any blueprint", "📜"),

		// Project achievements
		c.projectAchievement("first_project", "Project Manager", "Complete a project", "📦"),

		// Habit achievements
		c.habitAchievement("habit_former", "Habit Former", "Create a habit", "🔁"),
	}

	return achievements
}

// CountEarned returns how many achievements have been earned.
func (c *AchievementChecker) CountEarned() int {
	count := 0
	for _, a := range c.GetAchievements() {
		if a.Earned {
			count++
		}
	}
	return count
}

// CountTotal returns total number of achievements.
func (c *AchievementChecker) CountTotal() int {
	return len(c.GetAchievements())
}

func (c *AchievementChecker) levelAchievement(id, name, desc, icon string, level int) Achievement {
	earned := LevelForTotalXP(c.player.XPTotal) >= level
	return Achievement{ID: id, Name: name, Description: desc, Icon: icon, Earned: earned}
}

func (c *AchievementChecker) taskCountAchievement(id, name, desc, icon string, count int) Achievement {
	doneCount := 0
	for _, t := range c.tasks {
		if t.Status == "done" && !t.IsProject {
			doneCount++
		}
	}
	earned := doneCount >= count
	return Achievement{ID: id, Name: name, Description: desc, Icon: icon, Earned: earned}
}

func (c *AchievementChecker) attrLevelAchievement(id, name, desc, icon, attr string, level int) Achievement {
	attrXP := 0
	switch attr {
	case "str":
		attrXP = c.player.XPStr
	case "int":
		attrXP = c.player.XPInt
	case "wis":
		attrXP = c.player.XPWis
	case "art":
		attrXP = c.player.XPArt
	case "home":
		attrXP = c.player.XPHome
	case "out":
		attrXP = c.player.XPOut
	case "read":
		attrXP = c.player.XPRead
	case "cinema":
		attrXP = c.player.XPCinema
	case "career":
		attrXP = c.player.XPCareer
	}
	earned := AttributeLevelForXP(attrXP) >= level
	return Achievement{ID: id, Name: name, Description: desc, Icon: icon, Earned: earned}
}

func (c *AchievementChecker) blueprintAchievement(id, name, desc, icon string) Achievement {
	earned := false
	for _, b := range c.blueprints {
		if b.Status == "completed" {
			earned = true
			break
		}
	}
	return Achievement{ID: id, Name: name, Description: desc, Icon: icon, Earned: earned}
}

func (c *AchievementChecker) projectAchievement(id, name, desc, icon string) Achievement {
	earned := false
	for _, t := range c.tasks {
		if t.IsProject && t.Status == "done" {
			earned = true
			break
		}
	}
	return Achievement{ID: id, Name: name, Description: desc, Icon: icon, Earned: earned}
}

func (c *AchievementChecker) habitAchievement(id, name, desc, icon string) Achievement {
	earned := false
	for _, t := range c.tasks {
		if t.IsHabit {
			earned = true
			break
		}
	}
	return Achievement{ID: id, Name: name, Description: desc, Icon: icon, Earned: earned}
}

// GetAchievementsForPlayer is a convenience function.
func GetAchievementsForPlayer(ctx context.Context, svc *Service) ([]Achievement, error) {
	player, err := svc.PlayerRepo().GetOrCreateMain(ctx)
	if err != nil {
		return nil, err
	}
	tasks, err := svc.TaskRepo().ListAll(ctx)
	if err != nil {
		return nil, err
	}
	blueprints, err := svc.BlueprintRepo().ListAll(ctx)
	if err != nil {
		return nil, err
	}
	checker := NewAchievementChecker(player, tasks, blueprints)
	return checker.GetAchievements(), nil
}
