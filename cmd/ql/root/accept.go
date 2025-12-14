package root

import (
	"context"
	"errors"
	"fmt"

	"questline/internal/ui"

	"github.com/spf13/cobra"
)

func newAcceptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accept <blueprint_id>",
		Short: "Accept a blueprint and instantiate it",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("blueprint_id is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			code := args[0]
			svc, cleanup, err := openService(ctx)
			if err != nil {
				return err
			}
			defer cleanup()

			res, err := svc.AcceptBlueprint(ctx, code)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s → created #%d\n", ui.Good.Render(ui.IconScroll+" Accepted"), ui.Muted.Render(code), res.TaskID)
			return nil
		},
	}

	return cmd
}
