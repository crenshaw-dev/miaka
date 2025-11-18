package cmd

import (
	"fmt"
	"os"

	"github.com/crenshaw-dev/miaka/pkg/build/validation"
	"github.com/spf13/cobra"
)

var (
	validateCRDPath    string
	validateSchemaPath string
)

var validateCmd = &cobra.Command{
	Use:   "validate [values-file]",
	Short: "Validate a values file against CRD and JSON Schema",
	Long: `Validate a Helm values file against the generated CRD and JSON Schema.

This command helps chart maintainers test that actual values files from their
users pass the validation rules defined in the generated schemas.

The command validates against both:
  - Kubernetes CRD with OpenAPI v3 schema
  - JSON Schema for Helm validation

By default, the command looks for crd.yaml and values.schema.json in the
current directory.`,
	Example: `  # Validate values.yaml against default schemas
  miaka validate values.yaml

  # Validate with custom schema paths
  miaka validate values.yaml --crd output/crd.yaml --schema output/values.schema.json

  # Validate user-provided values
  miaka validate user-values.yaml`,
	Args:          cobra.ExactArgs(1),
	RunE:          runValidate,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	validateCmd.Flags().StringVarP(&validateCRDPath, "crd", "c", "crd.yaml", "Path to CRD YAML file")
	validateCmd.Flags().StringVarP(&validateSchemaPath, "schema", "s", "values.schema.json", "Path to JSON Schema file")
}

func runValidate(_ *cobra.Command, args []string) error {
	valuesPath := args[0]

	// Check that all required files exist
	if _, err := os.Stat(valuesPath); os.IsNotExist(err) {
		return fmt.Errorf("values file not found: %s", valuesPath)
	}
	if _, err := os.Stat(validateCRDPath); os.IsNotExist(err) {
		return fmt.Errorf("CRD file not found: %s", validateCRDPath)
	}
	if _, err := os.Stat(validateSchemaPath); os.IsNotExist(err) {
		return fmt.Errorf("JSON Schema file not found: %s", validateSchemaPath)
	}

	// Track validation results
	hasErrors := false

	// Validate against CRD
	fmt.Printf("Validating against CRD (%s)...\n", validateCRDPath)
	if err := validation.ValidateAgainstCRD(validateCRDPath, valuesPath); err != nil {
		fmt.Printf("✗ CRD validation failed: %v\n", err)
		hasErrors = true
	} else {
		fmt.Println("✓ CRD validation passed")
	}

	fmt.Println()

	// Validate against JSON Schema
	fmt.Printf("Validating against JSON Schema (%s)...\n", validateSchemaPath)
	if err := validation.ValidateYAML(valuesPath, validateSchemaPath); err != nil {
		fmt.Printf("✗ JSON Schema validation failed: %v\n", err)
		hasErrors = true
	} else {
		fmt.Println("✓ JSON Schema validation passed")
	}

	if hasErrors {
		return fmt.Errorf("validation failed")
	}

	return nil
}
