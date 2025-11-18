package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newInitCommand creates a fresh init command instance for testing
func newInitCommand() *cobra.Command {
	// Reset flags to defaults
	initApiVersion = ""
	initKind = ""
	initOutput = "example.values.yaml"

	// Create new command
	cmd := &cobra.Command{
		Use:   "init [values.yaml]",
		Short: "Convert values.yaml to KRM-compliant example.values.yaml",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runInit,
	}

	cmd.Flags().StringVar(&initApiVersion, "api-version", "", "API version (e.g., myapp.io/v1)")
	cmd.Flags().StringVar(&initKind, "kind", "", "Kind name (e.g., MyApp)")
	cmd.Flags().StringVarP(&initOutput, "output", "o", "example.values.yaml", "Output file path")

	return cmd
}

// TestInitCommand_BasicConversion tests basic conversion with all flags provided
func TestInitCommand_BasicConversion(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "values.yaml")
	outputPath := filepath.Join(tmpDir, "example.values.yaml")

	// Create input file
	inputContent := `replicaCount: 3
image: nginx:latest
service:
  port: 80
  type: ClusterIP
`
	if err := os.WriteFile(inputPath, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	// Run command
	cmd := newInitCommand()
	cmd.SetArgs([]string{
		inputPath,
		"--api-version", "myapp.io/v1",
		"--kind", "MyApp",
		"-o", outputPath,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Init command failed: %v\nStderr: %s", err, errBuf.String())
	}

	// Verify output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file was not created: %s", outputPath)
	}

	// Verify output content
	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "apiVersion: myapp.io/v1") {
		t.Error("Output does not contain apiVersion")
	}
	if !strings.Contains(outputStr, "kind: MyApp") {
		t.Error("Output does not contain kind")
	}
	if !strings.Contains(outputStr, "replicaCount: 3") {
		t.Error("Output does not contain original replicaCount field")
	}
}

// TestInitCommand_EmptyFile tests creating an empty KRM file
func TestInitCommand_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "example.values.yaml")

	// Change to tmpDir so that default values.yaml lookup doesn't find anything
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Run command without input file (will create empty KRM file)
	cmd := newInitCommand()
	cmd.SetArgs([]string{
		"--api-version", "test.io/v1",
		"--kind", "Test",
		"-o", outputPath,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Init command failed: %v\nStderr: %s", err, errBuf.String())
	}

	// Verify output file was created
	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "apiVersion: test.io/v1") {
		t.Error("Output does not contain apiVersion")
	}
	if !strings.Contains(outputStr, "kind: Test") {
		t.Error("Output does not contain kind")
	}
}

// TestInitCommand_MissingFlags tests that command fails when required flags are missing in non-interactive mode
func TestInitCommand_MissingFlags(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "values.yaml")
	outputPath := filepath.Join(tmpDir, "example.values.yaml")

	// Create input file without apiVersion/kind
	inputContent := `replicaCount: 3
image: nginx:latest
`
	if err := os.WriteFile(inputPath, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	// Test missing apiVersion
	cmd := newInitCommand()
	cmd.SetArgs([]string{
		inputPath,
		"--kind", "Test",
		"-o", outputPath,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error when apiVersion is missing, but command succeeded")
	}

	// Test missing kind
	cmd2 := newInitCommand()
	cmd2.SetArgs([]string{
		inputPath,
		"--api-version", "test.io/v1",
		"-o", outputPath,
	})

	outBuf2 := new(bytes.Buffer)
	errBuf2 := new(bytes.Buffer)
	cmd2.SetOut(outBuf2)
	cmd2.SetErr(errBuf2)

	err = cmd2.Execute()
	if err == nil {
		t.Fatal("Expected error when kind is missing, but command succeeded")
	}
}

// TestInitCommand_ExistingKRMFields tests behavior when input already has apiVersion/kind
func TestInitCommand_ExistingKRMFields(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "values.yaml")
	outputPath := filepath.Join(tmpDir, "example.values.yaml")

	// Create input file with apiVersion and kind
	inputContent := `apiVersion: v1
kind: ConfigMap
data:
  foo: bar
`
	if err := os.WriteFile(inputPath, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	// Run command without flags (should use existing values from file)
	cmd := newInitCommand()
	cmd.SetArgs([]string{
		inputPath,
		"-o", outputPath,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Init command failed: %v\nStderr: %s", err, errBuf.String())
	}

	// Verify output file preserves existing values
	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "apiVersion: v1") {
		t.Error("Output does not preserve existing apiVersion")
	}
	if !strings.Contains(outputStr, "kind: ConfigMap") {
		t.Error("Output does not preserve existing kind")
	}
}

// TestInitCommand_OnlyOneKRMField tests error handling when only apiVersion or kind is present
func TestInitCommand_OnlyOneKRMField(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "example.values.yaml")

	// Test with only apiVersion in file
	inputPath1 := filepath.Join(tmpDir, "only-api.yaml")
	inputContent1 := `apiVersion: v1
data:
  foo: bar
`
	if err := os.WriteFile(inputPath1, []byte(inputContent1), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	cmd1 := newInitCommand()
	cmd1.SetArgs([]string{
		inputPath1,
		"-o", outputPath,
	})

	outBuf1 := new(bytes.Buffer)
	errBuf1 := new(bytes.Buffer)
	cmd1.SetOut(outBuf1)
	cmd1.SetErr(errBuf1)

	err := cmd1.Execute()
	if err == nil {
		t.Error("Expected error when file has apiVersion but missing kind")
	} else if !strings.Contains(err.Error(), "kind") {
		t.Errorf("Expected error about missing kind, got: %v", err)
	}

	// Test with only kind in file
	inputPath2 := filepath.Join(tmpDir, "only-kind.yaml")
	inputContent2 := `kind: ConfigMap
data:
  foo: bar
`
	if err := os.WriteFile(inputPath2, []byte(inputContent2), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	cmd2 := newInitCommand()
	cmd2.SetArgs([]string{
		inputPath2,
		"-o", outputPath,
	})

	outBuf2 := new(bytes.Buffer)
	errBuf2 := new(bytes.Buffer)
	cmd2.SetOut(outBuf2)
	cmd2.SetErr(errBuf2)

	err = cmd2.Execute()
	if err == nil {
		t.Error("Expected error when file has kind but missing apiVersion")
	} else if !strings.Contains(err.Error(), "apiVersion") {
		t.Errorf("Expected error about missing apiVersion, got: %v", err)
	}
}

// TestInitCommand_NonExistentInputFile tests error handling for specified non-existent files
func TestInitCommand_NonExistentInputFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "example.values.yaml")

	// When user explicitly specifies a non-existent file, it should fail
	cmd := newInitCommand()
	cmd.SetArgs([]string{
		"nonexistent.yaml",
		"--api-version", "test.io/v1",
		"--kind", "Test",
		"-o", outputPath,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error for non-existent input file, but command succeeded")
	}

	if !strings.Contains(err.Error(), "failed to read") {
		t.Errorf("Expected 'failed to read' error, got: %v", err)
	}
}

// TestInitCommand_InvalidYAML tests error handling for invalid YAML input
func TestInitCommand_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "invalid.yaml")
	outputPath := filepath.Join(tmpDir, "example.values.yaml")

	// Create invalid YAML file
	invalidContent := `this is not: valid: yaml: content`
	if err := os.WriteFile(inputPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	cmd := newInitCommand()
	cmd.SetArgs([]string{
		inputPath,
		"--api-version", "test.io/v1",
		"--kind", "Test",
		"-o", outputPath,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error for invalid YAML, but command succeeded")
	}

	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

// TestInitCommand_CustomOutputPath tests custom output file path
func TestInitCommand_CustomOutputPath(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "values.yaml")
	customOutput := filepath.Join(tmpDir, "custom-output.yaml")

	// Create input file
	inputContent := `replicaCount: 3`
	if err := os.WriteFile(inputPath, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	cmd := newInitCommand()
	cmd.SetArgs([]string{
		inputPath,
		"--api-version", "test.io/v1",
		"--kind", "Test",
		"-o", customOutput,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Init command failed: %v", err)
	}

	// Verify custom output file was created
	if _, err := os.Stat(customOutput); os.IsNotExist(err) {
		t.Fatalf("Custom output file was not created: %s", customOutput)
	}

	// Verify content
	output, err := os.ReadFile(customOutput)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !strings.Contains(string(output), "apiVersion: test.io/v1") {
		t.Error("Output does not contain apiVersion")
	}
}

// TestInitCommand_CommentPreservation tests that comments are preserved
func TestInitCommand_CommentPreservation(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "values.yaml")
	outputPath := filepath.Join(tmpDir, "example.values.yaml")

	// Create input file with comments
	inputContent := `# Replica count
replicaCount: 3
# Docker image
image: nginx:latest
`
	if err := os.WriteFile(inputPath, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	cmd := newInitCommand()
	cmd.SetArgs([]string{
		inputPath,
		"--api-version", "test.io/v1",
		"--kind", "Test",
		"-o", outputPath,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Init command failed: %v", err)
	}

	// Verify comments are preserved
	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "# Replica count") {
		t.Error("Comment for replicaCount was not preserved")
	}
	if !strings.Contains(outputStr, "# Docker image") {
		t.Error("Comment for image was not preserved")
	}
}

// TestInitCommand_ComplexStructure tests handling of nested objects and arrays
func TestInitCommand_ComplexStructure(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "values.yaml")
	outputPath := filepath.Join(tmpDir, "example.values.yaml")

	// Create input with nested structures
	inputContent := `service:
  port: 80
  type: ClusterIP
  annotations:
    foo: bar
env:
- name: FOO
  value: bar
- name: BAZ
  value: qux
`
	if err := os.WriteFile(inputPath, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	cmd := newInitCommand()
	cmd.SetArgs([]string{
		inputPath,
		"--api-version", "test.io/v1",
		"--kind", "Test",
		"-o", outputPath,
	})

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Init command failed: %v", err)
	}

	// Verify nested structure is preserved
	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "service:") {
		t.Error("Nested service object not preserved")
	}
	if !strings.Contains(outputStr, "port: 80") {
		t.Error("Nested service.port not preserved")
	}
	if !strings.Contains(outputStr, "annotations:") {
		t.Error("Deeply nested annotations not preserved")
	}
	if !strings.Contains(outputStr, "env:") {
		t.Error("Array field not preserved")
	}
	if !strings.Contains(outputStr, "name: FOO") {
		t.Error("Array item not preserved")
	}
}
