package cmd

import (
	"fmt"

	initpkg "github.com/crenshaw-dev/miaka/pkg/init"
	"github.com/spf13/cobra"
)

var (
	initApiVersion string
	initKind       string
	initOutput     string
)

var initCmd = &cobra.Command{
	Use:   "init [values.yaml]",
	Short: "Convert values.yaml to KRM-compliant example.values.yaml",
	Long: `Convert a regular Helm values.yaml to KRM-compliant format.

This command adds apiVersion and kind fields to your YAML while preserving 
all existing comments and keeping all fields at the top level (not nested 
under spec). The metadata field is added automatically by the build command 
as a Kubernetes implementation detail. This is ideal for legacy Helm charts 
where you want minimal changes.

If apiVersion and kind already exist in the input file, they will be used 
and the flags are optional. If they don't exist, the flags are required.

If no input file is provided, an empty KRM-compliant YAML will be generated.`,
	Example: `  # Convert values.yaml to KRM format
  miaka init --api-version=myapp.io/v1 --kind=MyApp values.yaml

  # With custom output file
  miaka init --api-version=myapp.io/v1 --kind=MyApp -o custom.yaml values.yaml
  
  # Use existing apiVersion/kind from input file
  miaka init input.yaml  # (if input already has apiVersion and kind)
  
  # Generate empty KRM file (requires flags)
  miaka init --api-version=myapp.io/v1 --kind=MyApp -o example.values.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVar(&initApiVersion, "api-version", "", "API version (e.g., myapp.io/v1)")
	initCmd.Flags().StringVar(&initKind, "kind", "", "Kind name (e.g., MyApp)")
	initCmd.Flags().StringVarP(&initOutput, "output", "o", "example.values.yaml", "Output file path")

	// Don't mark as required - we'll validate conditionally in runInit
}

func runInit(cmd *cobra.Command, args []string) error {
	var inputFile string
	if len(args) > 0 {
		inputFile = args[0]
	}

	if err := initpkg.ConvertToKRM(inputFile, initOutput, initApiVersion, initKind); err != nil {
		return fmt.Errorf("failed to convert: %w", err)
	}

	if inputFile != "" {
		fmt.Printf("Successfully converted %s to %s\n", inputFile, initOutput)
	} else {
		fmt.Printf("Successfully created %s\n", initOutput)
	}
	return nil
}
