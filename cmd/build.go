// Package cmd provides the command-line interface for miaka.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/crenshaw-dev/miaka/pkg/build/generation/crd"
	"github.com/crenshaw-dev/miaka/pkg/build/generation/gotypes"
	"github.com/crenshaw-dev/miaka/pkg/build/generation/jsonschema"
	"github.com/crenshaw-dev/miaka/pkg/build/parsing"
	"github.com/crenshaw-dev/miaka/pkg/build/schema"
	"github.com/crenshaw-dev/miaka/pkg/build/validation"
	"github.com/spf13/cobra"
	runtimeschema "k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	defaultExampleValuesFile = "example.values.yaml"
	defaultCRDPath           = "crd.yaml"
	defaultSchemaPath        = "values.schema.json"
)

var (
	buildTypesPath  string
	buildCRDPath    string
	buildSchemaPath string
)

var buildCmd = &cobra.Command{
	Use:   "build [example.values.yaml]",
	Short: "Generate Go types and/or CRD from example.values.yaml",
	Long: `Generate Go types and Kubernetes Custom Resource Definitions (CRDs)
from a KRM-compliant YAML file.

If no input file is specified, the command will look for example.values.yaml 
in the current directory.

The tool generates:
  - Go type definitions (types.go) with proper struct tags
  - Kubernetes CRD with OpenAPI v3 schema (optional)

The generated CRD includes field descriptions, validation rules, and all
kubebuilder markers from your YAML comments.`,
	Example: `  # Generate CRD from example.values.yaml (default)
  miaka build

  # Generate from a specific file
  miaka build myfile.yaml

  # Generate CRD with custom path
  miaka build -c output/my-crd.yaml

  # Generate CRD and preserve types.go
  miaka build -t types.go

  # Custom types.go and CRD output locations
  miaka build -t pkg/apis/v1/types.go -c crds/my-crd.yaml myfile.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBuild,
	// SilenceUsage prevents usage from showing on business logic errors
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().StringVarP(&buildTypesPath, "types", "t", "", "Output path for types.go file (if empty, types.go is not preserved)")
	buildCmd.Flags().StringVarP(&buildCRDPath, "crd", "c", defaultCRDPath, "Output path for CRD YAML file")
	buildCmd.Flags().StringVarP(&buildSchemaPath, "schema", "s", defaultSchemaPath, "Output path for JSON Schema file")
}

func runBuild(_ *cobra.Command, args []string) error {
	// Determine input file: use provided arg, or default to example.values.yaml
	inputFile := defaultExampleValuesFile
	if len(args) > 0 {
		inputFile = args[0]
	}

	// Check if input file exists
	if _, err := os.Stat(inputFile); err != nil {
		if len(args) == 0 {
			return fmt.Errorf("%s not found in current directory (specify a file or run 'miaka init' first)", defaultExampleValuesFile)
		}
		return fmt.Errorf("input file not found: %s", inputFile)
	}

	// Parse the YAML file
	p := parsing.NewParser()
	s, err := p.ParseFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Prepare types file path and cleanup
	typesFilePath, cleanup, err := prepareTypesFile()
	if err != nil {
		return err
	}
	defer cleanup()

	// Generate and write types
	if err := generateAndWriteTypes(s, inputFile, typesFilePath); err != nil {
		return err
	}

	// Generate CRD with breaking change detection
	hadExistingCRD, err := handleCRDGeneration(s, typesFilePath, inputFile)
	if err != nil {
		return err
	}

	// Generate and validate JSON Schema
	if err := generateJSONSchema(inputFile); err != nil {
		return err
	}

	// Print next steps for first-time users
	if !hadExistingCRD {
		printNextSteps(inputFile)
	}

	return nil
}

// prepareTypesFile sets up the types file path and returns a cleanup function
func prepareTypesFile() (typesFilePath string, cleanup func(), err error) {
	if buildTypesPath != "" {
		// Use specified output path
		return buildTypesPath, func() {}, nil
	}

	// No --types flag specified, use temp file
	tmpDir, err := os.MkdirTemp("", "miaka-build-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	cleanup = func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to remove temp directory %s: %v\n", tmpDir, err)
		}
	}
	return filepath.Join(tmpDir, "types.go"), cleanup, nil
}

// generateAndWriteTypes generates Go types and writes them to file
func generateAndWriteTypes(s *schema.Schema, inputFile, typesFilePath string) error {
	fmt.Printf("Generating Go types from %s...\n", inputFile)
	g := gotypes.NewGenerator(s)
	code, err := g.Generate()

	// Write types.go file even if there were formatting errors (for debugging)
	if len(code) > 0 {
		if writeErr := os.WriteFile(typesFilePath, code, 0644); writeErr != nil {
			if err != nil {
				return fmt.Errorf("failed to write types file: %w (original error: %w)", writeErr, err)
			}
			return fmt.Errorf("failed to write types file: %w", writeErr)
		}
		if err != nil && buildTypesPath != "" {
			fmt.Fprintf(os.Stderr, "\nUnformatted types written to: %s\n", typesFilePath)
		}
	}

	// Return error if generation failed
	if err != nil {
		return fmt.Errorf("failed to generate Go code: %w", err)
	}

	fmt.Println("✓ Go types generated successfully")

	// Validate schema after writing types (so users can inspect the file on failure)
	fmt.Println("Validating schema...")
	if err := schema.ValidateSchema(s); err != nil {
		if buildTypesPath != "" {
			fmt.Fprintf(os.Stderr, "\nGenerated types with issues written to: %s\n", typesFilePath)
		}
		return err
	}

	fmt.Println("✓ Schema validation passed")

	// Print success message for types if preserving them
	if buildTypesPath != "" {
		fmt.Printf("✓ Types saved to %s\n", typesFilePath)
	}

	return nil
}

// handleCRDGeneration generates CRD and handles breaking change detection
func handleCRDGeneration(s *schema.Schema, typesFilePath, inputFile string) (hadExistingCRD bool, err error) {
	fmt.Printf("Generating CRD %s...\n", buildCRDPath)

	crdDir := filepath.Dir(buildCRDPath)
	crdFileName := filepath.Base(buildCRDPath)

	// Save existing CRD for breaking change comparison (if it exists)
	var oldCRDContent []byte
	if existingData, readErr := os.ReadFile(buildCRDPath); readErr == nil {
		oldCRDContent = existingData
		hadExistingCRD = true
		fmt.Println("Checking for breaking changes against existing CRD...")
	}

	if err := generateCRD(s, typesFilePath, crdDir, crdFileName); err != nil {
		return hadExistingCRD, fmt.Errorf("failed to generate CRD: %w", err)
	}

	// Check for breaking changes if there was an existing CRD
	if hadExistingCRD {
		if err := checkBreakingChanges(oldCRDContent); err != nil {
			return hadExistingCRD, err
		}
	}

	// Add strict validation to CRD (additionalProperties: false)
	if err := crd.AddStrictValidation(buildCRDPath); err != nil {
		return hadExistingCRD, fmt.Errorf("failed to add strict validation to CRD: %w", err)
	}

	// Validate the generated CRD itself
	if err := crd.ValidateCRD(buildCRDPath); err != nil {
		return hadExistingCRD, fmt.Errorf("generated CRD is invalid: %w", err)
	}

	fmt.Printf("✓ CRD generated: %s\n", buildCRDPath)

	// Validate the input YAML against the generated CRD
	fmt.Printf("Validating %s against CRD...\n", inputFile)
	if err := validation.ValidateAgainstCRD(buildCRDPath, inputFile); err != nil {
		return hadExistingCRD, fmt.Errorf("validation failed: %w", err)
	}

	fmt.Printf("✓ Validation passed: %s conforms to CRD schema\n", inputFile)

	return hadExistingCRD, nil
}

// checkBreakingChanges compares old and new CRD for breaking changes
func checkBreakingChanges(oldCRDContent []byte) error {
	newCRDContent, err := os.ReadFile(buildCRDPath)
	if err != nil {
		return fmt.Errorf("failed to read generated CRD: %w", err)
	}

	// Create temp file with old CRD for comparison
	tmpOldCRD, err := os.CreateTemp("", "old-crd-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file for old CRD: %w", err)
	}
	defer os.Remove(tmpOldCRD.Name())

	if _, err := tmpOldCRD.Write(oldCRDContent); err != nil {
		tmpOldCRD.Close()
		return fmt.Errorf("failed to write old CRD to temp file: %w", err)
	}
	tmpOldCRD.Close()

	// Check for breaking changes
	if err := validation.CheckBreakingChanges(tmpOldCRD.Name(), newCRDContent); err != nil {
		// Restore the old CRD since we're rejecting the breaking change
		if writeErr := os.WriteFile(buildCRDPath, oldCRDContent, 0644); writeErr != nil {
			return fmt.Errorf("breaking change detected and failed to restore old CRD: %w (original error: %w)", writeErr, err)
		}
		return fmt.Errorf("failed to generate CRD: %w", err)
	}

	return nil
}

// generateJSONSchema generates and validates JSON Schema
func generateJSONSchema(inputFile string) error {
	// Generate JSON Schema
	fmt.Printf("Generating JSON Schema %s...\n", buildSchemaPath)
	if err := jsonschema.GenerateFromCRD(buildCRDPath, buildSchemaPath); err != nil {
		return fmt.Errorf("failed to generate JSON Schema: %w", err)
	}
	fmt.Printf("✓ JSON Schema generated: %s\n", buildSchemaPath)

	// Validate input against JSON Schema
	fmt.Printf("Validating %s against JSON Schema...\n", inputFile)
	if err := validation.ValidateYAML(inputFile, buildSchemaPath); err != nil {
		return fmt.Errorf("JSON Schema validation failed: %w", err)
	}
	fmt.Printf("✓ JSON Schema validation passed\n")

	return nil
}

// printNextSteps prints helpful next steps for first-time users
func printNextSteps(inputFile string) {
	fmt.Println()
	fmt.Println("🎉 Generated schemas for the first time!")
	fmt.Println()
	fmt.Println("📝 Next steps:")
	fmt.Println("  1. Validate your actual values files:")
	fmt.Printf("       miaka validate your-values.yaml\n")
	fmt.Println()
	fmt.Println("  2. Improve your schema by editing", inputFile+":")
	fmt.Println("       - Add kubebuilder validation markers (e.g., +kubebuilder:validation:Minimum=1)")
	fmt.Println("       - Add field descriptions as comments")
	fmt.Println("       - Then run 'miaka build' again to regenerate schemas")
	fmt.Println()
	fmt.Println("  3. Commit the generated files to git:")
	fmt.Printf("       git add %s %s %s\n", buildCRDPath, buildSchemaPath, inputFile)
	fmt.Println("       git commit -m 'Add Miaka schemas'")
	fmt.Println("       (This enables breaking change detection on future builds)")
}

// generateCRD generates a CRD from the schema and types file
func generateCRD(s *schema.Schema, typesFile string, outputDir string, outputFileName string) error {
	// Parse apiVersion using Kubernetes libraries
	gv, err := runtimeschema.ParseGroupVersion(s.APIVersion)
	if err != nil {
		return fmt.Errorf("invalid apiVersion format: %s: %w", s.APIVersion, err)
	}

	// Create CRD generator options
	opts := crd.Options{
		Group:          gv.Group,
		Version:        gv.Version,
		Kind:           s.Kind,
		OutputFileName: outputFileName,
	}

	// Create CRD generator
	gen := crd.NewGenerator(opts)

	// Generate CRD
	return gen.Generate(typesFile, outputDir)
}
