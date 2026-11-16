package main

import (
	"github.com/JonasNowak/dms-gop/gops"
	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Show available commands",
	Long:  "Display all available dms-gop commands and options.",
	Run: func(cmd *cobra.Command, args []string) {
		printHeader()
		rootCmd.Usage()
	},
}

func runHelpCommand(gopsUtil *gops.GopsUtil) error {
	printHeader()
	rootCmd.Usage()
	return nil
}
