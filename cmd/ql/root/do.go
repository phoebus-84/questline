package root

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"questline/internal/ui"
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
			before, _ := svc.TaskRepo().Get(ctx, id)
			res, err := svc.CompleteTask(ctx, id)
			if err != nil {
				return err
			}

			name := fmt.Sprintf("#%d", res.TaskID)
			if before != nil {
				name = fmt.Sprintf("%s #%d %s", ui.KindIcon(before.IsProject, before.IsHabit), res.TaskID, before.Title)
			}
			line := fmt.Sprintf("%s %s %s", ui.Good.Render(ui.IconDone+" Completed"), name, ui.Muted.Render(fmt.Sprintf("(+%d XP)", res.XPAwarded)))
			fmt.Fprintln(cmd.OutOrStdout(), line)
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", ui.LabelValue("Level", fmt.Sprintf("%d → %d", res.LevelBefore, res.LevelAfter)))
			if res.ProjectBonus {
				fmt.Fprintln(cmd.OutOrStdout(), ui.Gold.Render(ui.IconTrophy+" Project bonus")+" "+ui.Muted.Render(fmt.Sprintf("(volume=%d)", res.ProjectVolume)))
			}
			if res.LevelUp {
				fmt.Fprintln(cmd.OutOrStdout(), ui.Gold.Render(ui.IconBolt+" "+ui.BadgeLevelUp))
			}
			return nil
		},
	}

	return cmd
}
