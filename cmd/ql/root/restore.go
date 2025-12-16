package root

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"questline/internal/ui"
)

func newRestoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore <id>",
		Short: "Restore a completed task (undo completion)",
		Long: `Restore a task to pending status by undoing its last completion.

This will:
- Remove the last completion record
- Deduct the XP that was awarded
- Reset the task status to pending

Use this to fix accidental completions or when a task was marked done prematurely.`,
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
			res, err := svc.RestoreTask(ctx, id)
			if err != nil {
				return err
			}

			name := fmt.Sprintf("#%d", res.TaskID)
			if before != nil {
				name = fmt.Sprintf("%s #%d %s", ui.KindIcon(before.IsProject, before.IsHabit), res.TaskID, before.Title)
			}
			line := fmt.Sprintf("%s %s %s", ui.Warn.Render(ui.IconUndo+" Restored"), name, ui.Muted.Render(fmt.Sprintf("(-%d XP)", res.XPDeducted)))
			fmt.Fprintln(cmd.OutOrStdout(), line)
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", ui.LabelValue("Level", fmt.Sprintf("%d → %d", res.LevelBefore, res.LevelAfter)))
			if res.LevelDown {
				fmt.Fprintln(cmd.OutOrStdout(), ui.Warn.Render(ui.IconWarn+" Level decreased"))
			}
			return nil
		},
	}

	return cmd
}
