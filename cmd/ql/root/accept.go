package root

import (
	"errors"

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
			return errors.New("not implemented (bootstrap)")
		},
	}

	return cmd
}
