package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var appVersion = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Показати версію",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(appVersion)
	},
}

func init() { rootCmd.AddCommand(versionCmd) }
