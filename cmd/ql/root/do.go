package root

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newDoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "do <id>",
		Short: "Complete a task",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("id is required")
			}
			if _, err := strconv.ParseInt(args[0], 10, 64); err != nil {
				return errors.New("id must be an integer")
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

			id, _ := strconv.ParseInt(args[0], 10, 64)
			res, err := svc.CompleteTask(ctx, id)
			if err != nil {
				return err
			}

			msg := fmt.Sprintf("Completed %d: +%d XP (level %d → %d)", res.TaskID, res.XPAwarded, res.LevelBefore, res.LevelAfter)
			if res.ProjectBonus {
				msg += fmt.Sprintf(" [project bonus, volume=%d]", res.ProjectVolume)
			}
			if res.LevelUp {
				msg += " [LEVEL UP]"
			}
			fmt.Fprintln(cmd.OutOrStdout(), msg)
			return nil
		},
	}

	return cmd
}
