package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func HelpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "help [command]",
		Short: "Display help for Syncora CLI or a specific command",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Root().Help()
				return
			}
			if subCmd, _, err := cmd.Root().Find(args); err == nil {
				subCmd.Help()
			} else {
				fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
				cmd.Root().Help()
			}
		},
	}
	return cmd
}