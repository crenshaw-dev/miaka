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
	tmpDir, err := os.MkdirTemp("", "crdgen-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

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
	err = os.WriteFile(typesFile, []byte(typesContent), 0644)
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
	assert.NoError(t, err, "Expected CRD file not found: %s", crdPath)

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
