package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vangjvn/devchain/version"
)

// versionCmd ...
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
	},
}
