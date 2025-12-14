package root

import (
	"errors"

	"github.com/spf13/cobra"
)

func newBoardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "board",
		Short: "Open the TUI dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("not implemented (bootstrap)")
		},
	}

	return cmd
}
