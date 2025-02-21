package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/vangjvn/devchain/sdk/client/commands/auto"
	basecmd "github.com/vangjvn/devchain/server/commands"
)

// TravisCmd is the entry point for this binary
var (
	TravisCmd = &cobra.Command{
		Use:   "devchain",
		Short: "Second State DevChain",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	lineBreak = &cobra.Command{Run: func(*cobra.Command, []string) {}}
)

func main() {
	// disable sorting
	cobra.EnableCommandSorting = false

	// add commands
	prepareNodeCommands()
	prepareClientCommands()
	prepareAccountCommands()

	TravisCmd.AddCommand(
		accountCmd,
		nodeCmd,
		clientCmd,
		attachCmd,
		basecmd.RemoveAddrBookCmd,
		basecmd.ResetPrivValidatorCmd,
		versionCmd,

		lineBreak,
		auto.AutoCompleteCmd,
	)

	// prepare and add flags
	basecmd.SetUpRoot(TravisCmd)
	executor := cli.PrepareMainCmd(TravisCmd, "CM", os.ExpandEnv("$HOME/.devchain"))
	executor.Execute()
}
