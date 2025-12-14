package root

import (
	"errors"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show player stats and unlocks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("not implemented (bootstrap)")
		},
	}

	return cmd
}
