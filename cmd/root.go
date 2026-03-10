package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	serverOverride string
	jsonOutput     bool
	quietOutput    bool
)

var rootCmd = &cobra.Command{
	Use:   "hctf2",
	Short: "hCTF2 — self-hosted CTF platform",
	Long:  "hCTF2 is a self-hosted CTF platform. Run 'hctf2 serve' to start the server.",
}

// Execute runs the root command with the given version string.
func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&serverOverride, "server", "", "Server URL (overrides config)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVar(&quietOutput, "quiet", false, "Minimal output")
}
