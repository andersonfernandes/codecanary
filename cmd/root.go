package cmd

import (
	"github.com/spf13/cobra"
)

var Version = "dev"

// DisplayVersion returns the version formatted for display.
// Release versions keep their "v" prefix (e.g. "v1.2.3").
// Non-release versions like "canary" or "dev" have any "v" prefix stripped.
func DisplayVersion() string {
	v := Version
	if len(v) > 1 && v[0] == 'v' && (v[1] < '0' || v[1] > '9') {
		v = v[1:]
	}
	return v
}

var rootCmd = &cobra.Command{
	Use:   "codecanary",
	Short: "AI-powered code review for GitHub pull requests",
	Long:  "Catch bugs, security issues, and quality problems before they land in main.",
}

func Execute() error {
	rootCmd.Version = DisplayVersion()
	return rootCmd.Execute()
}
