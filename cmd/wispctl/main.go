package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "wispctl",
	Short:         "Wispmail CLI management tool",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(tenantsCmd)
	rootCmd.AddCommand(usersCmd)
	rootCmd.AddCommand(domainsCmd)
	rootCmd.AddCommand(statsCmd)
}

var tenantsCmd = &cobra.Command{Use: "tenants", Short: "Manage tenants"}
var usersCmd = &cobra.Command{Use: "users", Short: "Manage users"}
var domainsCmd = &cobra.Command{Use: "domains", Short: "Manage domains"}
var statsCmd = &cobra.Command{Use: "stats", Short: "View system statistics"}