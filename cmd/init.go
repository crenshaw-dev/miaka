package cmd

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	initpkg "github.com/crenshaw-dev/miaka/pkg/init"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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

If no input file is specified, the command will look for values.yaml in the 
current directory. If values.yaml doesn't exist, an empty KRM-compliant YAML 
will be created.

If apiVersion and kind are not provided via flags and not present in the input 
file, the command will prompt you interactively for these values (unless running 
in non-interactive mode like CI/CD).`,
	Example: `  # Convert values.yaml to KRM format (will prompt for apiVersion/kind)
  miaka init

  # Provide apiVersion and kind via flags
  miaka init --api-version=myapp.io/v1 --kind=MyApp

  # Convert a different file
  miaka init --api-version=myapp.io/v1 --kind=MyApp myvalues.yaml

  # With custom output file
  miaka init --api-version=myapp.io/v1 --kind=MyApp -o custom.yaml
  
  # Use existing apiVersion/kind from input file
  miaka init input.yaml  # (if input already has apiVersion and kind)`,
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
	// Determine input file: use provided arg, or default to values.yaml
	inputFile := "values.yaml"
	if len(args) > 0 {
		inputFile = args[0]
	}

	// Check if input file exists
	fileExists := false
	if _, err := os.Stat(inputFile); err == nil {
		fileExists = true
	}

	// If default values.yaml doesn't exist and no arg was provided, treat as empty
	if !fileExists && len(args) == 0 {
		inputFile = ""
	}

	// Get apiVersion and kind, prompting interactively if needed
	apiVersion := initApiVersion
	kind := initKind

	// Check if the input file already has apiVersion and kind
	hasApiVersion, hasKind := false, false
	if fileExists {
		hasApiVersion, hasKind = initpkg.CheckKRMFields(inputFile)
	}

	// Only prompt if values not provided via flags, not in file, and we're in a TTY
	if term.IsTerminal(int(os.Stdin.Fd())) {
		if apiVersion == "" && !hasApiVersion {
			prompt := &survey.Input{
				Message: "API Version (e.g., myapp.io/v1):",
			}
			if err := survey.AskOne(prompt, &apiVersion, survey.WithValidator(survey.Required)); err != nil {
				return fmt.Errorf("failed to get API version: %w", err)
			}
		}

		if kind == "" && !hasKind {
			prompt := &survey.Input{
				Message: "Kind (e.g., MyApp):",
			}
			if err := survey.AskOne(prompt, &kind, survey.WithValidator(survey.Required)); err != nil {
				return fmt.Errorf("failed to get kind: %w", err)
			}
		}
	}

	if err := initpkg.ConvertToKRM(inputFile, initOutput, apiVersion, kind); err != nil {
		return fmt.Errorf("failed to convert: %w", err)
	}

	if inputFile != "" {
		fmt.Printf("Successfully converted %s to %s\n", inputFile, initOutput)
	} else {
		fmt.Printf("Successfully created %s\n", initOutput)
	}
	return nil
}
