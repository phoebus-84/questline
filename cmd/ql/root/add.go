package root

import (
	"context"
	"errors"
	"fmt"

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
			attrParsed := engine.ParseAttribute(attr)

			var parent *int64
			if parentID != 0 {
				v := parentID
				parent = &v
			}

			if isProject {
				res, err := svc.CreateProject(ctx, engine.CreateProjectInput{Title: title, Attribute: attrParsed})
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
			if isHabit {
				parsed, err := engine.ParseHabitInterval(habitInterval)
				if err != nil {
					return err
				}
				interval = parsed
			}

			res, err := svc.CreateTask(ctx, engine.CreateTaskInput{
				Title:         title,
				Difficulty:    d,
				Attribute:     attrParsed,
				ParentID:      parent,
				IsHabit:       isHabit,
				HabitInterval: interval,
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
			fmt.Fprintln(cmd.OutOrStdout(), line)
			return nil
		},
	}

	cmd.Flags().IntVarP(&diff, "diff", "d", 1, "Difficulty (1-5)")
	cmd.Flags().StringVarP(&attr, "attr", "a", "wis", "Attribute (str|int|wis|art)")
	cmd.Flags().Int64VarP(&parentID, "parent", "p", 0, "Parent task ID (subtasks/projects)")
	cmd.Flags().BoolVar(&isProject, "project", false, "Create a project container (requires unlock)")
	cmd.Flags().BoolVar(&isHabit, "habit", false, "Create a recurring habit (requires unlock)")
	cmd.Flags().StringVar(&habitInterval, "interval", "daily", "Habit interval (daily|weekly)")

	return cmd
}
