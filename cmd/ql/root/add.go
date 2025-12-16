package root

import (
	"context"
	"errors"
	"fmt"
	"time"

	"questline/internal/engine"
	"questline/internal/ui"

	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var diff int
	var attr string
	var parentID int64
	var isProject bool
	var isHabit bool
	var habitInterval string
	var habitDuration string
	var habitGoal int

	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Add a task (or project/habit)",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("title is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			svc, cleanup, err := openService(ctx)
			if err != nil {
				return err
			}
			defer cleanup()

			title := args[0]
			primaryAttr, attrWeights := engine.ParseAttributes(attr)

			var parent *int64
			if parentID != 0 {
				v := parentID
				parent = &v
			}

			if isProject {
				res, err := svc.CreateProject(ctx, engine.CreateProjectInput{
					Title:      title,
					Attribute:  primaryAttr,
					Attributes: attrWeights,
				})
				if err != nil {
					return err
				}
				created, _ := svc.TaskRepo().Get(ctx, res.TaskID)
				fmt.Fprintln(cmd.OutOrStdout(), ui.Good.Render(ui.IconBox+" Created project")+" "+fmt.Sprintf("#%d %s", res.TaskID, created.Title)+" "+ui.Muted.Render("("+ui.StatusText(created.Status)+")"))
				return nil
			}

			if diff < 1 || diff > 5 {
				return errors.New("diff must be between 1 and 5")
			}
			d := engine.Difficulty(diff)

			var interval engine.HabitInterval
			var duration *time.Duration
			var goal *int
			if isHabit {
				parsed, err := engine.ParseHabitInterval(habitInterval)
				if err != nil {
					return err
				}
				interval = parsed

				// Parse duration (e.g., "7d", "1w", "30d", "1m")
				if habitDuration != "" {
					dur, err := parseDuration(habitDuration)
					if err != nil {
						return fmt.Errorf("invalid duration: %w", err)
					}
					duration = &dur
				}

				// Set goal
				if habitGoal > 0 {
					goal = &habitGoal
				}
			}

			res, err := svc.CreateTask(ctx, engine.CreateTaskInput{
				Title:         title,
				Difficulty:    d,
				Attribute:     primaryAttr,
				Attributes:    attrWeights,
				ParentID:      parent,
				IsHabit:       isHabit,
				HabitInterval: interval,
				HabitDuration: duration,
				HabitGoal:     goal,
			})
			if err != nil {
				return err
			}
			created, err := svc.TaskRepo().Get(ctx, res.TaskID)
			if err != nil {
				return err
			}

			icon := ui.KindIcon(created.IsProject, created.IsHabit)
			label := "Created task"
			if created.IsHabit {
				label = "Created habit"
			}

			line := ui.Good.Render(ui.IconPlus+" "+label) + " " + fmt.Sprintf("%s #%d %s", icon, res.TaskID, created.Title)
			line += " " + ui.Muted.Render(fmt.Sprintf("(+%d XP)", created.XPValue))
			if res.ProjectActivated {
				line += " " + ui.Gold.Render("⚡ project activated")
			}
			if goal != nil {
				line += " " + ui.Muted.Render(fmt.Sprintf("[0/%d]", *goal))
			}
			fmt.Fprintln(cmd.OutOrStdout(), line)
			return nil
		},
	}

	cmd.Flags().IntVarP(&diff, "diff", "d", 1, "Difficulty (1-5)")
	cmd.Flags().StringVarP(&attr, "attr", "a", "wis", "Attribute(s): single (str) or multi (str:50,int:50)")
	cmd.Flags().Int64VarP(&parentID, "parent", "p", 0, "Parent task ID (subtasks/projects)")
	cmd.Flags().BoolVar(&isProject, "project", false, "Create a project container (requires unlock)")
	cmd.Flags().BoolVar(&isHabit, "habit", false, "Create a recurring habit (requires unlock)")
	cmd.Flags().StringVar(&habitInterval, "interval", "daily", "Habit interval (daily|weekly|monthly)")
	cmd.Flags().StringVar(&habitDuration, "duration", "", "Habit duration (e.g., 7d, 1w, 30d, 1m)")
	cmd.Flags().IntVar(&habitGoal, "goal", 0, "Target completions to finish the habit")

	return cmd
}

// parseDuration parses a duration string like "7d", "1w", "30d", "1m"
func parseDuration(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, fmt.Errorf("duration too short: %s", s)
	}
	unit := s[len(s)-1]
	numStr := s[:len(s)-1]
	var num int
	if _, err := fmt.Sscanf(numStr, "%d", &num); err != nil {
		return 0, fmt.Errorf("invalid number in duration: %s", s)
	}
	if num <= 0 {
		return 0, fmt.Errorf("duration must be positive: %s", s)
	}

	switch unit {
	case 'd':
		return time.Duration(num) * 24 * time.Hour, nil
	case 'w':
		return time.Duration(num) * 7 * 24 * time.Hour, nil
	case 'm':
		return time.Duration(num) * 30 * 24 * time.Hour, nil // Approximate month
	default:
		return 0, fmt.Errorf("unknown duration unit: %c (use d/w/m)", unit)
	}
}
