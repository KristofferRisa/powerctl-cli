package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is the current version of powerctl-cli
	// This can be overridden at build time using -ldflags
	Version = "dev"
	// GitCommit is the git commit hash (set at build time)
	GitCommit = "unknown"
	// BuildDate is when the binary was built (set at build time)
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Display the current version, git commit, and build date of powerctl-cli.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("powerctl-cli version %s\n", Version)
		fmt.Printf("Git commit: %s\n", GitCommit)
		fmt.Printf("Built: %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
