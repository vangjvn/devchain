package main

import (
	basecmd "github.com/vangjvn/devchain/server/commands"
	"github.com/spf13/cobra"
)

// nodeCmd is the entry point for this binary
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Second State DevChain",
	Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func prepareNodeCommands() {
	nodeCmd.AddCommand(
		basecmd.InitCmd,
		basecmd.GetStartCmd(),
		basecmd.ShowNodeIDCmd,
	)
}
