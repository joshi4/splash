package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of splash",
	Long:  `Print the version number of splash`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("splash %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
