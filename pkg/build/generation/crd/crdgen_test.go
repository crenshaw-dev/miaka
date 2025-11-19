package crd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerator_Generate tests the CRD generation with temp directory isolation
func TestGenerator_Generate(t *testing.T) {
	// Create temp directory for test output
	tmpDir := t.TempDir()

	// Create a sample types.go file
	typesContent := `package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
//
// Example is the Schema for the examples API
type Example struct {
	metav1.TypeMeta   ` + "`json:\",inline\"`" + `
	metav1.ObjectMeta ` + "`json:\"metadata,omitempty\"`" + `

	Spec ExampleSpec ` + "`json:\"spec,omitempty\"`" + `
}

// ExampleSpec defines the desired state of Example
type ExampleSpec struct {
	// Replicas is the number of desired pods
	Replicas int ` + "`json:\"replicas,omitempty\"`" + `
}
`

	typesFile := filepath.Join(tmpDir, "types.go")
	err := os.WriteFile(typesFile, []byte(typesContent), 0644)
	require.NoError(t, err, "Failed to write types file")

	opts := Options{
		Group:          "example.com",
		Version:        "v1",
		Kind:           "Example",
		OutputFileName: "test-crd.yaml",
	}

	gen := NewGenerator(opts)

	// Generate CRD - all intermediate files should be in temp, not polluting tmpDir
	outputDir := filepath.Join(tmpDir, "output")
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err, "Failed to create output dir")

	err = gen.Generate(typesFile, outputDir)
	require.NoError(t, err, "Generate failed")

	// Verify CRD was created
	crdPath := filepath.Join(outputDir, "test-crd.yaml")
	_, err = os.Stat(crdPath)
	require.NoError(t, err, "Expected CRD file not found: %s", crdPath)

	// Verify CRD content
	crdContent, err := os.ReadFile(crdPath)
	require.NoError(t, err, "Failed to read CRD")

	crdStr := string(crdContent)
	assert.Contains(t, crdStr, "apiVersion: apiextensions.k8s.io/v1", "CRD missing apiVersion")
	assert.Contains(t, crdStr, "kind: CustomResourceDefinition", "CRD missing kind")
	assert.Contains(t, crdStr, "example.com", "CRD missing group")

	// Verify no intermediate files were left in the types file directory
	// (They should all be in a temp directory that was cleaned up)
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err, "Failed to read tmpDir")

	for _, entry := range entries {
		name := entry.Name()
		// Only types.go and output dir should exist, no go.mod, go.sum, or doc.go
		assert.True(t, name == "types.go" || name == "output",
			"Unexpected file in tmpDir: %s (intermediate files should be cleaned up)", name)
	}
}

// TestOptions tests the Options struct
func TestOptions(t *testing.T) {
	opts := Options{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	}

	assert.Equal(t, "apps", opts.Group)
	assert.Equal(t, "v1", opts.Version)
}

// TestGenerator_Generate_DefaultOutputFilename tests that controller-gen generates
// predictable filenames and we can find them without a fallback
func TestGenerator_Generate_DefaultOutputFilename(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a sample types.go file
	typesContent := `package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
//
// MyApp is the Schema for the myapps API
type MyApp struct {
	metav1.TypeMeta   ` + "`json:\",inline\"`" + `
	metav1.ObjectMeta ` + "`json:\"metadata,omitempty\"`" + `

	Spec MyAppSpec ` + "`json:\"spec,omitempty\"`" + `
}

// MyAppSpec defines the desired state of MyApp
type MyAppSpec struct {
	Replicas int ` + "`json:\"replicas,omitempty\"`" + `
}
`

	typesFile := filepath.Join(tmpDir, "types.go")
	err := os.WriteFile(typesFile, []byte(typesContent), 0644)
	require.NoError(t, err, "Failed to write types file")

	opts := Options{
		Group:   "myapp.io",
		Version: "v1alpha1",
		Kind:    "MyApp",
		// No OutputFileName specified - let controller-gen choose
	}

	gen := NewGenerator(opts)

	outputDir := filepath.Join(tmpDir, "output")
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err, "Failed to create output dir")

	err = gen.Generate(typesFile, outputDir)
	require.NoError(t, err, "Generate failed")

	// List what controller-gen actually created
	entries, err := os.ReadDir(outputDir)
	require.NoError(t, err, "Failed to read output dir")

	assert.Len(t, entries, 1, "Expected exactly 1 file in output dir")

	generatedFile := entries[0].Name()
	t.Logf("Controller-gen generated file: %s", generatedFile)

	// Verify it matches the expected pattern
	// Controller-gen generates: <group>_<plural>.yaml
	expectedName := "myapp.io_myapps.yaml"
	assert.Equal(t, expectedName, generatedFile)
}
