package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configFile string
	version    = "dev"
	commit     = "unknown"
	buildTime  = "unknown"
)

var rootCmd = &cobra.Command{
	Use:           "wispmail",
	Short:         "Wispmail - Distributed Email Server",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Path to configuration file")
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(versionCmd)
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}