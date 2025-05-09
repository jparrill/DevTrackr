package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "devtrackr",
	Short: "DevTrackr - Track your Jira issues and pull requests",
	Long: `DevTrackr is a tool designed to help developers track their daily work
by monitoring Jira issues and their associated pull requests. It provides
a centralized way to manage and track the progress of features and bug fixes
across different releases.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
