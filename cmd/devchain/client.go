package main

import (
	"github.com/spf13/cobra"

	stakecmd "github.com/vangjvn/devchain/modules/stake/commands"
	"github.com/vangjvn/devchain/sdk/client/commands"
	"github.com/vangjvn/devchain/sdk/client/commands/query"
	txcmd "github.com/vangjvn/devchain/sdk/client/commands/txs"
)

// clientCmd is the entry point for this binary
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Travis light client",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func prepareClientCommands() {
	commands.AddBasicFlags(clientCmd)

	query.RootCmd.AddCommand(
		stakecmd.CmdQueryValidator,
		stakecmd.CmdQueryValidators,
	)

	// set up the middleware
	txcmd.Middleware = txcmd.Wrappers{}
	txcmd.Middleware.Register(txcmd.RootCmd.PersistentFlags())

	txcmd.RootCmd.AddCommand(
		stakecmd.CmdDeclareCandidacy,
		stakecmd.CmdUpdateCandidacy,
		stakecmd.CmdWithdrawCandidacy,
		stakecmd.CmdVerifyCandidacy,
		stakecmd.CmdActivateCandidacy,
		stakecmd.CmdDeactivateCandidacy,
		stakecmd.CmdUpdateCandidacyAccount,
		stakecmd.CmdAcceptCandidacyAccountUpdate,
	)

	clientCmd.AddCommand(
		txcmd.RootCmd,
		query.RootCmd,
		lineBreak,

		commands.InitCmd,
		commands.ResetCmd,
	)
}
