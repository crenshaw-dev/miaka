package crd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// Options contains configuration for CRD generation
type Options struct {
	// Group is the API group (e.g., "example.com")
	Group string

	// Version is the API version (e.g., "v1alpha1")
	Version string

	// Kind is the CRD kind (e.g., "Example")
	Kind string

	// OutputFileName is the name of the output CRD file
	// If empty, defaults to <group>_<version>_<kind>.yaml
	OutputFileName string
}

// Generator generates CRDs using controller-gen as a library
type Generator struct {
	opts Options
}

// NewGenerator creates a new CRD generator
func NewGenerator(opts Options) *Generator {
	return &Generator{
		opts: opts,
	}
}

// Generate creates a CRD YAML file from a types.go file using controller-gen library
// All intermediate files are created in a temporary directory to avoid polluting the user's filesystem
func (g *Generator) Generate(typesFile string, outputDir string) error {
	// Validate inputs
	if _, err := os.Stat(typesFile); err != nil {
		return fmt.Errorf("types file not found: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create temporary directory for all intermediate files
	tmpDir, err := os.MkdirTemp("", "crdgen-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy types.go to temp directory
	tmpTypesFile := filepath.Join(tmpDir, "types.go")
	if err := copyFile(typesFile, tmpTypesFile); err != nil {
		return fmt.Errorf("failed to copy types file: %w", err)
	}

	// Create go.mod in temp directory
	goModPath := filepath.Join(tmpDir, "go.mod")
	goModContent := fmt.Sprintf("module generated\n\ngo 1.21\n\nrequire k8s.io/apimachinery v0.31.0\n")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}

	// Create doc.go with package-level markers in temp directory
	docGoPath := filepath.Join(tmpDir, "doc.go")
	docGoContent := fmt.Sprintf(`// +kubebuilder:object:generate=true
// +groupName=%s
package %s
`, g.opts.Group, g.opts.Version)
	if err := os.WriteFile(docGoPath, []byte(docGoContent), 0644); err != nil {
		return fmt.Errorf("failed to create doc.go: %w", err)
	}

	// Run go mod tidy in temp directory to generate go.sum
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run go mod tidy: %w\nOutput: %s", err, string(output))
	}

	// Create a subdirectory in temp for controller-gen output
	tmpOutputDir := filepath.Join(tmpDir, "output")
	if err := os.MkdirAll(tmpOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp output directory: %w", err)
	}

	// Use the genall framework like controller-gen CLI does
	// Build the options as if they were command-line arguments
	options := []string{
		"crd:crdVersions=v1",
		"paths=" + tmpDir,
		"output:crd:dir=" + tmpOutputDir,
	}

	// Create options registry (same as controller-gen)
	optionsRegistry := &markers.Registry{}

	// Register the CRD generator as a marker
	crdGenDef := markers.Must(markers.MakeDefinition("crd", markers.DescribesPackage, crd.Generator{}))
	if err := optionsRegistry.Register(crdGenDef); err != nil {
		return fmt.Errorf("failed to register crd generator: %w", err)
	}

	// Register output rules
	outputDirRule := markers.Must(markers.MakeDefinition("output:crd:dir", markers.DescribesPackage, genall.OutputToDirectory("")))
	if err := optionsRegistry.Register(outputDirRule); err != nil {
		return fmt.Errorf("failed to register output rule: %w", err)
	}

	// Register common options (paths, etc)
	if err := genall.RegisterOptionsMarkers(optionsRegistry); err != nil {
		return fmt.Errorf("failed to register options markers: %w", err)
	}

	// Create runtime from options (like controller-gen does)
	rt, err := genall.FromOptions(optionsRegistry, options)
	if err != nil {
		return fmt.Errorf("failed to create runtime from options: %w", err)
	}

	// Run the generators
	if hadErrs := rt.Run(); hadErrs {
		return fmt.Errorf("CRD generation failed - controller-gen encountered errors processing the generated types (run with --keep-types to inspect the generated code)")
	}

	// Find the generated CRD file in temp output directory
	generatedCRDPath, err := findCRDFile(tmpOutputDir, g.opts.Group, g.opts.Kind)
	if err != nil {
		return fmt.Errorf("failed to find generated CRD: %w", err)
	}

	// Determine final output filename
	finalFileName := g.opts.OutputFileName
	if finalFileName == "" {
		// Use the generated filename
		finalFileName = filepath.Base(generatedCRDPath)
	}
	finalOutputPath := filepath.Join(outputDir, finalFileName)

	// Copy the generated CRD to the user's output directory
	if err := copyFile(generatedCRDPath, finalOutputPath); err != nil {
		return fmt.Errorf("failed to copy CRD to output directory: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

// findCRDFile finds the generated CRD file in a directory
// Returns the path to the CRD file or an error if not found
func findCRDFile(dir, group, kind string) (string, error) {
	// Controller-gen generates files with naming convention: <group>_<plural>.yaml
	// We need to find the file that matches this pattern
	plural := strings.ToLower(kind) + "s"
	expectedName := fmt.Sprintf("%s_%s.yaml", group, plural)
	expectedPath := filepath.Join(dir, expectedName)

	if _, err := os.Stat(expectedPath); err == nil {
		return expectedPath, nil
	}

	// If not found, scan the directory for any .yaml files
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("failed to read output directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			return filepath.Join(dir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("no CRD file found in %s (expected %s)", dir, expectedName)
}
