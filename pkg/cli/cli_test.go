package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {
	// Save original stdout and create a buffer
	oldStdout := rootCmd.OutOrStdout()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	defer rootCmd.SetOut(oldStdout)

	// Execute version command
	rootCmd.SetArgs([]string{"version"})
	err := rootCmd.Execute()

	// Verify
	assert.NoError(t, err)
	output := buf.String()
	t.Logf("Command output: %q", output)
	assert.Contains(t, output, "DevTrackr version")
}

func TestRootCommand(t *testing.T) {
	// Test that root command exists and has correct properties
	assert.Equal(t, "devtrackr", rootCmd.Use)
	assert.Contains(t, rootCmd.Short, "Track your Jira issues and pull requests")
	assert.Contains(t, rootCmd.Long, "DevTrackr is a tool designed to help developers")
}

func TestRootCmdFunction(t *testing.T) {
	// Test that RootCmd() returns the correct command
	cmd := RootCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, rootCmd, cmd)
}

func TestCommandStructure(t *testing.T) {
	// Test that version command is properly registered
	var versionCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "version" {
			versionCmd = cmd
			break
		}
	}
	assert.NotNil(t, versionCmd, "Version command not found")
	assert.Equal(t, "version", versionCmd.Use)
	assert.Contains(t, versionCmd.Short, "Print the version number")
}

func TestExecute(t *testing.T) {
	// Test that Execute doesn't panic with invalid command
	rootCmd.SetArgs([]string{"nonexistent"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}
