package root

import (
	"errors"

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
			return errors.New("not implemented (bootstrap)")
		},
	}

	cmd.Flags().IntVarP(&diff, "diff", "d", 1, "Difficulty (1-5)")
	cmd.Flags().StringVarP(&attr, "attr", "a", "wis", "Attribute (str|int|wis|art)")
	cmd.Flags().Int64VarP(&parentID, "parent", "p", 0, "Parent task ID (subtasks/projects)")
	cmd.Flags().BoolVar(&isProject, "project", false, "Create a project container (requires unlock)")
	cmd.Flags().BoolVar(&isHabit, "habit", false, "Create a recurring habit (requires unlock)")
	cmd.Flags().StringVar(&habitInterval, "interval", "daily", "Habit interval (daily|weekly)")

	_ = diff
	_ = attr
	_ = parentID
	_ = isProject
	_ = isHabit
	_ = habitInterval

	return cmd
}
