package root

import (
	"context"

	"github.com/spf13/cobra"

	"questline/internal/tui"
)

func newBoardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "board",
		Short: "Open the TUI dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			svc, cleanup, err := openService(ctx)
			if err != nil {
				return err
			}
			defer cleanup()

			return tui.RunBoard(ctx, svc, cmd.OutOrStdout())
		},
	}

	return cmd
}
