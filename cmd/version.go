package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var appVersion string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the application version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version:", appVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
