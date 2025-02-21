package main

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/vangjvn/devchain/console"
)

var (
	attachCmd = &cobra.Command{
		RunE:  remoteConsole,
		Use:   "attach",
		Short: "Start an interactive JavaScript environment (connect to node)",
	}
)

func remoteConsole(cmd *cobra.Command, args []string) error {
	if len(args) != 1 || len(args[0]) == 0 {
		return errors.New("missing url")
	}

	client, err := dialRPC(args[0])
	if err != nil {
		utils.Fatalf("Unable to attach to remote node: %v", err)
	}
	config := console.Config{
		DataDir: "",
		DocRoot: "",
		Client:  client,
		Preload: []string{},
	}

	console, err := console.New(config)
	if err != nil {
		utils.Fatalf("Failed to start the JavaScript console: %v", err)
	}
	defer console.Stop(false)

	// Otherwise print the welcome screen and enter interactive mode
	console.Welcome()
	console.Interactive()

	return nil
}

func dialRPC(endpoint string) (*rpc.Client, error) {
	if strings.HasPrefix(endpoint, "rpc:") || strings.HasPrefix(endpoint, "ipc:") {
		// Backwards compatibility with geth < 1.5 which required
		// these prefixes.
		endpoint = endpoint[4:]
	}
	return rpc.Dial(endpoint)
}
