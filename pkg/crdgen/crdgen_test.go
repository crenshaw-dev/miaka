package crdgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGenerator_Generate tests the CRD generation with temp directory isolation
func TestGenerator_Generate(t *testing.T) {
	// Create temp directory for test output
	tmpDir, err := os.MkdirTemp("", "crdgen-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
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
	if err := os.WriteFile(typesFile, []byte(typesContent), 0644); err != nil {
		t.Fatalf("Failed to write types file: %v", err)
	}

	opts := Options{
		Group:          "example.com",
		Version:        "v1",
		Kind:           "Example",
		OutputFileName: "test-crd.yaml",
	}

	gen := NewGenerator(opts)

	// Generate CRD - all intermediate files should be in temp, not polluting tmpDir
	outputDir := filepath.Join(tmpDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	err = gen.Generate(typesFile, outputDir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify CRD was created
	crdPath := filepath.Join(outputDir, "test-crd.yaml")
	if _, err := os.Stat(crdPath); os.IsNotExist(err) {
		t.Errorf("Expected CRD file not found: %s", crdPath)
	}

	// Verify CRD content
	crdContent, err := os.ReadFile(crdPath)
	if err != nil {
		t.Fatalf("Failed to read CRD: %v", err)
	}

	crdStr := string(crdContent)
	if !strings.Contains(crdStr, "apiVersion: apiextensions.k8s.io/v1") {
		t.Errorf("CRD missing apiVersion")
	}
	if !strings.Contains(crdStr, "kind: CustomResourceDefinition") {
		t.Errorf("CRD missing kind")
	}
	if !strings.Contains(crdStr, "example.com") {
		t.Errorf("CRD missing group")
	}

	// Verify no intermediate files were left in the types file directory
	// (They should all be in a temp directory that was cleaned up)
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read tmpDir: %v", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		// Only types.go and output dir should exist, no go.mod, go.sum, or doc.go
		if name != "types.go" && name != "output" {
			t.Errorf("Unexpected file in tmpDir: %s (intermediate files should be cleaned up)", name)
		}
	}
}

// TestOptions tests the Options struct
func TestOptions(t *testing.T) {
	opts := Options{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	}

	if opts.Group != "apps" {
		t.Errorf("Group = %s, want apps", opts.Group)
	}
	if opts.Version != "v1" {
		t.Errorf("Version = %s, want v1", opts.Version)
	}
}
