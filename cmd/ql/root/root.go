package root

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"questline/internal/ui"
)

const Version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:           "ql",
	Short:         "Questline (MVP) — local-first RPG task manager",
	Long:          "Questline is a local-first CLI/TUI task manager with RPG progression mechanics.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("{{.Name}} v{{.Version}}\n")

	rootCmd.AddCommand(
		newAddCmd(),
		newDoCmd(),
		newRestoreCmd(),
		newListCmd(),
		newStatusCmd(),
		newAcceptCmd(),
		newBoardCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, ui.Bad.Render(ui.IconError+" "+err.Error()))
		os.Exit(1)
	}
}
