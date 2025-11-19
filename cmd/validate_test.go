package cmd

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestValidateCommand_Testdata runs table-driven tests for all test cases in testdata/validate/
func TestValidateCommand_Testdata(t *testing.T) {
	testdataDir := "../testdata/validate"

	// Find all test case directories
	entries, err := os.ReadDir(testdataDir)
	require.NoError(t, err, "Failed to read testdata directory: %v")

	var testCases []struct {
		name        string
		dir         string
		expectError bool
		skip        bool
		skipFile    string
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		testCaseName := entry.Name()
		testCaseDir := filepath.Join(testdataDir, testCaseName)

		// Check if this directory has a values.yaml
		valuesPath := filepath.Join(testCaseDir, "values.yaml")
		if _, err := os.Stat(valuesPath); os.IsNotExist(err) {
			continue
		}

		// Check if this test should be skipped
		skipPath := filepath.Join(testCaseDir, ".skip")
		skipReason := ""
		shouldSkip := false
		if skipData, err := os.ReadFile(skipPath); err == nil {
			shouldSkip = true
			skipReason = string(skipData)
		}

		// Determine if we expect an error based on the test name
		expectError := strings.Contains(testCaseName, "invalid")

		testCases = append(testCases, struct {
			name        string
			dir         string
			expectError bool
			skip        bool
			skipFile    string
		}{
			name:        testCaseName,
			dir:         testCaseDir,
			expectError: expectError,
			skip:        shouldSkip,
			skipFile:    skipReason,
		})
	}

	if len(testCases) == 0 {
		t.Fatal("No test cases found in testdata/validate/")
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skipf("Test skipped: %s", tc.skipFile)
			}

			// Run the validate command
			err := runValidateTestCase(t, tc.dir)

			// Check error expectation
			if tc.expectError && err == nil {
				t.Fatal("Expected validation to fail, but it succeeded")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Expected validation to succeed, but it failed: %v", err)
			}
		})
	}
}

// runValidateTestCase executes the validate command for a test case and validates behavior
func runValidateTestCase(t *testing.T, testDir string) error {
	t.Helper()

	valuesPath := filepath.Join(testDir, "values.yaml")
	crdPath := filepath.Join(testDir, "crd.yaml")
	schemaPath := filepath.Join(testDir, "schema.json")

	// Create a new validate command for this test
	cmd := &cobra.Command{
		Use:  "validate",
		RunE: runValidate,
	}

	// Set flags
	validateCRDPath = crdPath
	validateSchemaPath = schemaPath

	// Set args
	cmd.SetArgs([]string{valuesPath})

	// Run the command - we just care about the error status, not the output
	// because fmt.Printf doesn't respect cobra's SetOut in our implementation
	err := cmd.Execute()

	return err
}

// TestValidateCommand_MissingFiles tests validation with missing CRD or schema files
func TestValidateCommand_MissingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid values file
	valuesPath := filepath.Join(tmpDir, "values.yaml")
	valuesContent := `apiVersion: example.com/v1alpha1
kind: Example
replicas: 3
`
	if err := os.WriteFile(valuesPath, []byte(valuesContent), 0644); err != nil {
		t.Fatalf("Failed to write values file: %v", err)
	}

	t.Run("MissingCRD", func(t *testing.T) {
		var stdout bytes.Buffer
		cmd := &cobra.Command{
			Use:  "validate",
			RunE: runValidate,
		}
		cmd.SetOut(&stdout)
		cmd.SetErr(&stdout)

		// Point to non-existent CRD
		validateCRDPath = filepath.Join(tmpDir, "nonexistent.yaml")
		validateSchemaPath = filepath.Join(tmpDir, "schema.json")
		cmd.SetArgs([]string{valuesPath})

		err := cmd.Execute()
		require.Error(t, err, "Expected error for missing CRD, got nil")
		assert.Contains(t, err.Error(), "CRD file not found", "Expected 'CRD file not found' error, got: %v")
	})

	t.Run("MissingSchema", func(t *testing.T) {
		// Create a minimal CRD
		crdPath := filepath.Join(tmpDir, "crd.yaml")
		crdContent := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: examples.example.com
spec:
  group: example.com
  names:
    kind: Example
    plural: examples
  scope: Namespaced
  versions:
  - name: v1alpha1
    schema:
      openAPIV3Schema:
        type: object
    served: true
    storage: true
`
		if err := os.WriteFile(crdPath, []byte(crdContent), 0644); err != nil {
			t.Fatalf("Failed to write CRD file: %v", err)
		}

		var stdout bytes.Buffer
		cmd := &cobra.Command{
			Use:  "validate",
			RunE: runValidate,
		}
		cmd.SetOut(&stdout)
		cmd.SetErr(&stdout)

		validateCRDPath = crdPath
		validateSchemaPath = filepath.Join(tmpDir, "nonexistent.json")
		cmd.SetArgs([]string{valuesPath})

		err := cmd.Execute()
		require.Error(t, err, "Expected error for missing schema, got nil")
		assert.Contains(t, err.Error(), "JSON Schema file not found", "Expected 'JSON Schema file not found' error, got: %v")
	})
}

// TestValidateCommand_InvalidYAML tests validation with malformed values file
func TestValidateCommand_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a malformed values file
	valuesPath := filepath.Join(tmpDir, "values.yaml")
	invalidYAML := `apiVersion: example.com/v1alpha1
kind: Example
this is not: valid: yaml: content
`
	if err := os.WriteFile(valuesPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write values file: %v", err)
	}

	// Create minimal CRD and schema
	crdPath := filepath.Join(tmpDir, "crd.yaml")
	crdContent := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: examples.example.com
spec:
  group: example.com
  names:
    kind: Example
    plural: examples
  scope: Namespaced
  versions:
  - name: v1alpha1
    schema:
      openAPIV3Schema:
        type: object
    served: true
    storage: true
`
	if err := os.WriteFile(crdPath, []byte(crdContent), 0644); err != nil {
		t.Fatalf("Failed to write CRD file: %v", err)
	}

	schemaPath := filepath.Join(tmpDir, "schema.json")
	schemaContent := `{"type": "object"}`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	var stdout bytes.Buffer
	cmd := &cobra.Command{
		Use:  "validate",
		RunE: runValidate,
	}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)

	validateCRDPath = crdPath
	validateSchemaPath = schemaPath
	cmd.SetArgs([]string{valuesPath})

	err := cmd.Execute()
	require.Error(t, err, "Expected error for invalid YAML, got nil")
	// The error should come from either CRD or schema validation failing to parse
}

// TestValidateCommand_NoArgs tests validation without providing values file
func TestValidateCommand_NoArgs(t *testing.T) {
	var stdout bytes.Buffer
	cmd := &cobra.Command{
		Use:  "validate",
		Args: cobra.ExactArgs(1),
		RunE: runValidate,
	}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)

	validateCRDPath = "crd.yaml"
	validateSchemaPath = "schema.json"
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err, "Expected error for no arguments, got nil")
	assert.Contains(t, err.Error(), "accepts 1 arg", "Expected 'accepts 1 arg' error, got: %v")
}
