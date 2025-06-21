package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/xilverfang/syncora/cmd/bridge/internal/commands"
)


func main()  {
	rootCmd := cobra.Command{
		Use: "syncora",
		Short: "Syncora CLI for aggregating blockchain bridge services",
		Long: `Syncora is a command-line tool to interact with 
		multiple blockchain bridge services, allowing users to check supported bridges, 
		networks, and execute token bridging operations.`,
	}

	rootCmd.AddCommand(commands.AccountCmd())
	rootCmd.AddCommand(commands.InfoCmd())
	rootCmd.AddCommand(commands.HelpCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}