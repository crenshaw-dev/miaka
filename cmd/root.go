package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "miaka",
	Short: "Generate Go types and CRDs from KRM-compliant YAML",
	Long: `miaka helps convert Helm charts to KRM-compliant golden charts.

It provides two main workflows:
  1. init  - Convert existing values.yaml to KRM-compliant format
  2. build - Generate Go types and CRDs with OpenAPI schema

For more information about golden charts and the two-step rendering process,
see the documentation at https://github.com/crenshaw-dev/miaka`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(initCmd)
}

