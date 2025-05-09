package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "1.0.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of DevTrackr",
	Long:  `Display the version number of DevTrackr and exit.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "DevTrackr version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
