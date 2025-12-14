package root

import (
	"errors"
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
			return errors.New("not implemented (bootstrap)")
		},
	}

	return cmd
}
