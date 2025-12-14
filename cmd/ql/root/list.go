package root

import (
	"errors"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks (tree view)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("not implemented (bootstrap)")
		},
	}

	return cmd
}
