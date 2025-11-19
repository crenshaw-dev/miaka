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

// TestBuildCommand_Testdata runs table-driven tests for all test cases in testdata/build/
func TestBuildCommand_Testdata(t *testing.T) {
	testdataDir := "../testdata/build"

	// Find all test case directories
	entries, err := os.ReadDir(testdataDir)
	require.NoError(t, err, "Failed to read testdata directory: %v")

	var testCases []struct {
		name     string
		dir      string
		skip     bool
		skipFile string
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		testCaseName := entry.Name()
		testCaseDir := filepath.Join(testdataDir, testCaseName)

		// Check if this directory has an input.yaml
		inputPath := filepath.Join(testCaseDir, "input.yaml")
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
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

		testCases = append(testCases, struct {
			name     string
			dir      string
			skip     bool
			skipFile string
		}{
			name:     testCaseName,
			dir:      testCaseDir,
			skip:     shouldSkip,
			skipFile: skipReason,
		})
	}

	if len(testCases) == 0 {
		t.Fatal("No test cases found in testdata/build/")
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skipf("Test case marked as skipped: %s", strings.TrimSpace(tc.skipFile))
			}

			runTestCase(t, tc.dir)
		})
	}
}

func runTestCase(t *testing.T, testCaseDir string) {
	// Create temp directory for output
	tmpDir := t.TempDir()
	typesOutput := filepath.Join(tmpDir, "types.go")
	crdOutput := filepath.Join(tmpDir, "crd.yaml")
	schemaOutput := filepath.Join(tmpDir, "schema.json")
	inputPath := filepath.Join(testCaseDir, "input.yaml")

	// Create a new command instance for this test
	cmd := newBuildCommand()
	cmd.SetArgs([]string{
		inputPath,
		"-t", typesOutput,
		"-c", crdOutput,
		"-s", schemaOutput,
	})

	// Capture output
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	// Execute command
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Build command failed: %v\nStderr: %s\nStdout: %s", err, errBuf.String(), outBuf.String())
	}

	// Verify types.go was created
	if _, err := os.Stat(typesOutput); os.IsNotExist(err) {
		t.Errorf("Expected types file not found: %s", typesOutput)
	}

	// Verify CRD was created
	if _, err := os.Stat(crdOutput); os.IsNotExist(err) {
		t.Errorf("Expected CRD file not found: %s", crdOutput)
	}

	// Verify JSON Schema was created
	if _, err := os.Stat(schemaOutput); os.IsNotExist(err) {
		t.Errorf("Expected JSON Schema file not found: %s", schemaOutput)
	}

	// Compare generated types with expected (if expected file exists)
	expectedTypesPath := filepath.Join(testCaseDir, "expected_types.go")
	if _, err := os.Stat(expectedTypesPath); err == nil {
		compareFiles(t, "types.go", typesOutput, expectedTypesPath)
	}

	// Compare generated CRD with expected (if expected file exists)
	expectedCRDPath := filepath.Join(testCaseDir, "expected_crd.yaml")
	if _, err := os.Stat(expectedCRDPath); err == nil {
		compareFiles(t, "crd.yaml", crdOutput, expectedCRDPath)
	}

	// Compare generated JSON Schema with expected (if expected file exists)
	expectedSchemaPath := filepath.Join(testCaseDir, "expected_schema.json")
	if _, err := os.Stat(expectedSchemaPath); err == nil {
		compareFiles(t, "schema.json", schemaOutput, expectedSchemaPath)
	}
}

// newBuildCommand creates a fresh build command instance for testing
func newBuildCommand() *cobra.Command {
	// Reset flags to defaults
	buildTypesPath = ""
	buildCRDPath = "crd.yaml"
	buildSchemaPath = "values.schema.json"

	// Create new command
	cmd := &cobra.Command{
		Use:          "build [example.yaml]",
		Short:        "Generate Go types and/or CRD from example.values.yaml",
		Args:         cobra.MaximumNArgs(1),
		RunE:         runBuild,
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&buildTypesPath, "types", "t", "", "Output path for types.go file (if empty, types.go is not preserved)")
	cmd.Flags().StringVarP(&buildCRDPath, "crd", "c", "crd.yaml", "Output path for CRD YAML file")
	cmd.Flags().StringVarP(&buildSchemaPath, "schema", "s", "values.schema.json", "Output path for JSON Schema file")

	return cmd
}

// TestBuildCommand_ValidationError tests that validation errors are reported correctly
func TestBuildCommand_ValidationError(t *testing.T) {
	// Create a test YAML with interface{} types
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "invalid.yaml")
	crdOutput := filepath.Join(tmpDir, "crd.yaml")
	schemaOutput := filepath.Join(tmpDir, "schema.json")

	invalidYAML := `apiVersion: test.io/v1
kind: Test
emptyArray: []
`
	if err := os.WriteFile(inputPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cmd := newBuildCommand()
	cmd.SetArgs([]string{
		inputPath,
		"-c", crdOutput,
		"-s", schemaOutput,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	require.Error(t, err, "Expected validation error but command succeeded")

	// Check error message contains validation information
	errOutput := err.Error()
	assert.Contains(t, errOutput, "interface{}", "Expected interface{} validation error, got: %s")
}

// TestBuildCommand_TypeHints tests that type hints work correctly
func TestBuildCommand_TypeHints(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "hints.yaml")
	typesOutput := filepath.Join(tmpDir, "types.go")
	crdOutput := filepath.Join(tmpDir, "crd.yaml")
	schemaOutput := filepath.Join(tmpDir, "schema.json")

	yamlWithHints := `apiVersion: test.io/v1
kind: Test
# +miaka:type:map[string]string
labels: {}
# +miaka:type:string
names: []
`
	if err := os.WriteFile(inputPath, []byte(yamlWithHints), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cmd := newBuildCommand()
	cmd.SetArgs([]string{
		inputPath,
		"-t", typesOutput,
		"-c", crdOutput,
		"-s", schemaOutput,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Build command failed: %v\nStderr: %s", err, errBuf.String())
	}

	// Read generated types
	content, err := os.ReadFile(typesOutput)
	require.NoError(t, err, "Failed to read generated types: %v")

	contentStr := string(content)

	// Verify type hints were applied
	assert.Contains(t, contentStr, "Labels map[string]string", "Expected 'Labels map[string]string' in generated types, got:\n%s")
	assert.Contains(t, contentStr, "Names []string", "Expected 'Names []string' in generated types, got:\n%s")
}

// TestBuildCommand_InvalidInput tests error handling for invalid input
func TestBuildCommand_InvalidInput(t *testing.T) {
	tmpDir := t.TempDir()
	crdOutput := filepath.Join(tmpDir, "crd.yaml")
	schemaOutput := filepath.Join(tmpDir, "schema.json")

	cmd := newBuildCommand()
	cmd.SetArgs([]string{
		"nonexistent.yaml",
		"-c", crdOutput,
		"-s", schemaOutput,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	require.Error(t, err, "Expected error for nonexistent file but command succeeded")

	if !strings.Contains(err.Error(), "failed to parse YAML") &&
		!strings.Contains(err.Error(), "no such file") &&
		!strings.Contains(err.Error(), "input file not found") {
		t.Errorf("Expected file error, got: %v", err)
	}
}

// TestBuildCommand_BreakingChangeDetection tests that breaking changes are detected and fail the build
func TestBuildCommand_BreakingChangeDetection(t *testing.T) {
	tmpDir := t.TempDir()
	crdOutput := filepath.Join(tmpDir, "crd.yaml")
	schemaOutput := filepath.Join(tmpDir, "schema.json")

	// Step 1: Create initial input and build CRD
	initialInput := filepath.Join(tmpDir, "initial.yaml")
	initialYAML := `apiVersion: example.com/v1
kind: Example
# Number of replicas
# +kubebuilder:validation:Minimum=1
replicas: 3
# Application name
appName: "myapp"
# Configuration settings
config:
  # Timeout in seconds
  timeout: 30
  # Retry count
  retries: 5
`
	if err := os.WriteFile(initialInput, []byte(initialYAML), 0644); err != nil {
		t.Fatalf("Failed to write initial input: %v", err)
	}

	// Build the initial CRD
	cmd := newBuildCommand()
	cmd.SetArgs([]string{
		initialInput,
		"-c", crdOutput,
		"-s", schemaOutput,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Initial build failed: %v\nStderr: %s\nStdout: %s", err, errBuf.String(), outBuf.String())
	}

	// Verify CRD was created
	if _, err := os.Stat(crdOutput); os.IsNotExist(err) {
		t.Fatalf("Initial CRD was not created: %s", crdOutput)
	}

	// Step 2: Create breaking change input (change field type from int to string)
	breakingInput := filepath.Join(tmpDir, "breaking.yaml")
	breakingYAML := `apiVersion: example.com/v1
kind: Example
# Number of replicas (changed to string - BREAKING CHANGE!)
replicas: "3"
# Application name
appName: "myapp"
# Configuration settings
config:
  # Timeout in seconds
  timeout: 30
  # Retry count
  retries: 5
`
	if err := os.WriteFile(breakingInput, []byte(breakingYAML), 0644); err != nil {
		t.Fatalf("Failed to write breaking input: %v", err)
	}

	// Step 3: Attempt to build with breaking change - should fail
	cmd2 := newBuildCommand()
	cmd2.SetArgs([]string{
		breakingInput,
		"-c", crdOutput,
		"-s", schemaOutput,
	})

	outBuf2 := new(bytes.Buffer)
	errBuf2 := new(bytes.Buffer)
	cmd2.SetOut(outBuf2)
	cmd2.SetErr(errBuf2)

	err = cmd2.Execute()
	require.Error(t, err, "Expected build to fail with breaking change detection, but it succeeded")

	// Verify error message mentions breaking changes
	errOutput := err.Error()
	assert.Contains(t, errOutput, "breaking changes detected", "Expected 'breaking changes detected' in error message, got: %s")

	t.Logf("Breaking change correctly detected: %s", errOutput)
}

// TestBuildCommand_CompatibleChange tests that compatible changes don't fail the build
func TestBuildCommand_CompatibleChange(t *testing.T) {
	tmpDir := t.TempDir()
	crdOutput := filepath.Join(tmpDir, "crd.yaml")
	schemaOutput := filepath.Join(tmpDir, "schema.json")

	// Step 1: Create initial input and build CRD
	initialInput := filepath.Join(tmpDir, "initial.yaml")
	initialYAML := `apiVersion: example.com/v1
kind: Example
# Number of replicas
replicas: 3
# Application name
appName: "myapp"
`
	if err := os.WriteFile(initialInput, []byte(initialYAML), 0644); err != nil {
		t.Fatalf("Failed to write initial input: %v", err)
	}

	// Build the initial CRD
	cmd := newBuildCommand()
	cmd.SetArgs([]string{
		initialInput,
		"-c", crdOutput,
		"-s", schemaOutput,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Initial build failed: %v\nStderr: %s", err, errBuf.String())
	}

	// Step 2: Create compatible change input (add a new optional field)
	compatibleInput := filepath.Join(tmpDir, "compatible.yaml")
	compatibleYAML := `apiVersion: example.com/v1
kind: Example
# Number of replicas
replicas: 3
# Application name
appName: "myapp"
# New optional field (compatible change)
description: "A description"
`
	if err := os.WriteFile(compatibleInput, []byte(compatibleYAML), 0644); err != nil {
		t.Fatalf("Failed to write compatible input: %v", err)
	}

	// Step 3: Build with compatible change - should succeed
	cmd2 := newBuildCommand()
	cmd2.SetArgs([]string{
		compatibleInput,
		"-c", crdOutput,
		"-s", schemaOutput,
	})

	outBuf2 := new(bytes.Buffer)
	errBuf2 := new(bytes.Buffer)
	cmd2.SetOut(outBuf2)
	cmd2.SetErr(errBuf2)

	err = cmd2.Execute()
	if err != nil {
		t.Fatalf("Build with compatible change should succeed, but failed: %v\nStderr: %s", err, errBuf2.String())
	}

	t.Log("Compatible change correctly allowed")
}

// compareFiles compares two files and reports differences
func compareFiles(t *testing.T, fileType, generatedPath, expectedPath string) {
	t.Helper()

	generated, err := os.ReadFile(generatedPath)
	require.NoError(t, err, "Failed to read generated %s: %v")

	expected, err := os.ReadFile(expectedPath)
	require.NoError(t, err, "Failed to read expected %s: %v")

	generatedStr := normalizeOutput(string(generated))
	expectedStr := normalizeOutput(string(expected))

	if generatedStr != expectedStr {
		t.Errorf("%s mismatch:\n\nFirst difference:\n%s",
			fileType,
			findFirstDifference(expectedStr, generatedStr),
		)
	}
}

// normalizeOutput normalizes whitespace and line endings for comparison
func normalizeOutput(s string) string {
	// Normalize line endings
	s = strings.ReplaceAll(s, "\r\n", "\n")

	// Normalize controller-gen version annotation (non-deterministic)
	s = normalizeControllerGenVersion(s)

	// Trim trailing whitespace from each line
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

// normalizeControllerGenVersion replaces the controller-gen version with a stable placeholder
func normalizeControllerGenVersion(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if strings.Contains(line, "controller-gen.kubebuilder.io/version:") {
			indent := strings.TrimRight(line[:strings.Index(line, "controller-gen")], " ")
			lines[i] = indent + "controller-gen.kubebuilder.io/version: <normalized>"
		}
	}
	return strings.Join(lines, "\n")
}

// findFirstDifference finds and formats the first difference between two strings
func findFirstDifference(expected, generated string) string {
	expLines := strings.Split(expected, "\n")
	genLines := strings.Split(generated, "\n")

	maxLines := len(expLines)
	if len(genLines) > maxLines {
		maxLines = len(genLines)
	}

	for i := 0; i < maxLines; i++ {
		var expLine, genLine string
		if i < len(expLines) {
			expLine = expLines[i]
		}
		if i < len(genLines) {
			genLine = genLines[i]
		}

		if expLine != genLine {
			return formatDiffLine(i+1, expLine, genLine)
		}
	}

	return "Files differ but no line-by-line difference found"
}

func formatDiffLine(lineNum int, expected, generated string) string {
	var sb strings.Builder
	sb.WriteString("Line ")
	sb.WriteString(intToString(lineNum))
	sb.WriteString(":\n")
	sb.WriteString("- Expected: ")
	sb.WriteString(expected)
	sb.WriteString("\n")
	sb.WriteString("+ Generated: ")
	sb.WriteString(generated)
	return sb.String()
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []rune{}
	for n > 0 {
		digits = append([]rune{rune(n%10 + '0')}, digits...)
		n /= 10
	}
	return string(digits)
}

// TestBuildCommand_MissingExampleValuesYaml tests error when example.values.yaml is missing
func TestBuildCommand_MissingExampleValuesYaml(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory (where example.values.yaml doesn't exist)
	origDir, err := os.Getwd()
	require.NoError(t, err, "Failed to get working directory: %v")
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	cmd := newBuildCommand()
	cmd.SetArgs([]string{}) // No arguments, should look for example.values.yaml

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err = cmd.Execute()
	require.Error(t, err, "Expected error for missing example.values.yaml but command succeeded")

	assert.Contains(t, err.Error(), "example.values.yaml not found", "Expected 'example.values.yaml not found' error, got: %v")
}

// TestBuildCommand_InvalidYaml tests error handling for malformed YAML
func TestBuildCommand_InvalidYaml(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "invalid.yaml")
	crdOutput := filepath.Join(tmpDir, "crd.yaml")
	schemaOutput := filepath.Join(tmpDir, "schema.json")

	// Write malformed YAML
	invalidYAML := `apiVersion: test.io/v1
kind: Test
this is not: valid: yaml: syntax
replicas: 3
`
	if err := os.WriteFile(inputPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cmd := newBuildCommand()
	cmd.SetArgs([]string{
		inputPath,
		"-c", crdOutput,
		"-s", schemaOutput,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	require.Error(t, err, "Expected error for invalid YAML but command succeeded")

	assert.Contains(t, err.Error(), "failed to parse YAML", "Expected 'failed to parse YAML' error, got: %v")
}

// TestBuildCommand_SaveTypesGo tests that types.go can be saved with -t flag
func TestBuildCommand_SaveTypesGo(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.yaml")
	typesOutput := filepath.Join(tmpDir, "saved-types.go")
	crdOutput := filepath.Join(tmpDir, "crd.yaml")
	schemaOutput := filepath.Join(tmpDir, "schema.json")

	validYAML := `apiVersion: example.com/v1
kind: Example
# Number of replicas
replicas: 3
# Application name
appName: "myapp"
`
	if err := os.WriteFile(inputPath, []byte(validYAML), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cmd := newBuildCommand()
	cmd.SetArgs([]string{
		inputPath,
		"-t", typesOutput,
		"-c", crdOutput,
		"-s", schemaOutput,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Build command failed: %v\nStderr: %s", err, errBuf.String())
	}

	// Verify types.go was created at the specified location
	if _, err := os.Stat(typesOutput); os.IsNotExist(err) {
		t.Errorf("Expected types file not found at specified location: %s", typesOutput)
	}

	// Verify the file contains expected content
	content, err := os.ReadFile(typesOutput)
	require.NoError(t, err, "Failed to read types file: %v")

	contentStr := string(content)
	assert.Contains(t, contentStr, "type Example struct", "Expected 'type Example struct' in types.go, got:\n%s")
	assert.Contains(t, contentStr, "Replicas", "Expected 'Replicas' field in types.go, got:\n%s")
	assert.Contains(t, contentStr, "AppName", "Expected 'AppName' field in types.go, got:\n%s")

	// Verify success message mentions the file (it prints to stdout via fmt.Printf, not cmd.OutOrStdout)
	// Since we're calling runBuild directly through cobra, check the actual console output isn't captured
	// Instead, let's just verify the file was created successfully
	t.Logf("Types file successfully created at: %s", typesOutput)
}

// TestBuildCommand_InvalidApiVersion tests error handling for invalid apiVersion format
func TestBuildCommand_InvalidApiVersion(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "invalid-version.yaml")
	crdOutput := filepath.Join(tmpDir, "crd.yaml")
	schemaOutput := filepath.Join(tmpDir, "schema.json")

	// Write YAML with invalid apiVersion format (malformed)
	invalidVersionYAML := `apiVersion: this/is/invalid/format
kind: Test
replicas: 3
`
	if err := os.WriteFile(inputPath, []byte(invalidVersionYAML), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cmd := newBuildCommand()
	cmd.SetArgs([]string{
		inputPath,
		"-c", crdOutput,
		"-s", schemaOutput,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	require.Error(t, err, "Expected error for invalid apiVersion format but command succeeded")

	errMsg := err.Error()
	if !strings.Contains(errMsg, "invalid apiVersion") && !strings.Contains(errMsg, "apiVersion") {
		t.Errorf("Expected apiVersion error, got: %v", err)
	}
}

// TestBuildCommand_BreakingChangeRestoresOldCRD tests that CRD is restored on breaking change
func TestBuildCommand_BreakingChangeRestoresOldCRD(t *testing.T) {
	tmpDir := t.TempDir()
	crdOutput := filepath.Join(tmpDir, "crd.yaml")
	schemaOutput := filepath.Join(tmpDir, "schema.json")

	// Step 1: Create and build initial CRD
	initialInput := filepath.Join(tmpDir, "initial.yaml")
	initialYAML := `apiVersion: example.com/v1
kind: Example
replicas: 3
`
	if err := os.WriteFile(initialInput, []byte(initialYAML), 0644); err != nil {
		t.Fatalf("Failed to write initial input: %v", err)
	}

	cmd := newBuildCommand()
	cmd.SetArgs([]string{
		initialInput,
		"-c", crdOutput,
		"-s", schemaOutput,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	require.NoError(t, err, "Initial build failed: %v")

	// Read the original CRD content
	originalCRD, err := os.ReadFile(crdOutput)
	require.NoError(t, err, "Failed to read original CRD: %v")

	// Step 2: Attempt breaking change
	breakingInput := filepath.Join(tmpDir, "breaking.yaml")
	breakingYAML := `apiVersion: example.com/v1
kind: Example
replicas: "3"
`
	if err := os.WriteFile(breakingInput, []byte(breakingYAML), 0644); err != nil {
		t.Fatalf("Failed to write breaking input: %v", err)
	}

	cmd2 := newBuildCommand()
	cmd2.SetArgs([]string{
		breakingInput,
		"-c", crdOutput,
		"-s", schemaOutput,
	})

	outBuf2 := new(bytes.Buffer)
	errBuf2 := new(bytes.Buffer)
	cmd2.SetOut(outBuf2)
	cmd2.SetErr(errBuf2)

	err = cmd2.Execute()
	require.Error(t, err, "Expected build to fail with breaking change")

	// Step 3: Verify the CRD was restored to original
	restoredCRD, err := os.ReadFile(crdOutput)
	require.NoError(t, err, "Failed to read restored CRD: %v")

	if !bytes.Equal(originalCRD, restoredCRD) {
		t.Error("CRD was not restored to original after breaking change detection")
	}
}
