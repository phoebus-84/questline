package root

import (
	"context"
	"errors"
	"fmt"

	"questline/internal/engine"
	"questline/internal/storage"

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

			path, err := storage.ResolveDBPath()
			if err != nil {
				return err
			}
			db, err := storage.Open(ctx, path)
			if err != nil {
				return err
			}
			defer db.Close()

			svc := engine.NewService(db)
			res, err := svc.AcceptBlueprint(ctx, code)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Accepted %s → created task %d\n", code, res.TaskID)
			return nil
		},
	}

	return cmd
}
