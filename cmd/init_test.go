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

// TestInitCommand_Testdata runs table-driven tests for all init test cases in testdata/init/
func TestInitCommand_Testdata(t *testing.T) {
	testdataDir := "../testdata/init"

	// Find all test case directories
	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Fatalf("Failed to read testdata/init directory: %v", err)
	}

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

		// Check if this directory has an expected.yaml
		expectedPath := filepath.Join(testCaseDir, "expected.yaml")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
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
		t.Fatal("No test cases found in testdata/init/")
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skipf("Test case marked as skipped: %s", strings.TrimSpace(tc.skipFile))
			}

			runInitTestCase(t, tc.dir)
		})
	}
}

func runInitTestCase(t *testing.T, testCaseDir string) {
	// Create temp directory for output
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.yaml")

	// Check if input file exists (may not exist for empty-file test)
	inputPath := filepath.Join(testCaseDir, "input.yaml")
	hasInput := false
	if _, err := os.Stat(inputPath); err == nil {
		hasInput = true
	}

	// Read flags from flags.txt (if present)
	var flags []string
	flagsPath := filepath.Join(testCaseDir, "flags.txt")
	if flagData, err := os.ReadFile(flagsPath); err == nil {
		flagLines := strings.Split(strings.TrimSpace(string(flagData)), "\n")
		for _, line := range flagLines {
			line = strings.TrimSpace(line)
			if line != "" {
				flags = append(flags, line)
			}
		}
	}

	// Build command arguments
	args := []string{}
	if hasInput {
		args = append(args, inputPath)
	}
	args = append(args, "-o", outputPath)
	args = append(args, flags...)

	// Create command
	cmd := newInitCommand()
	cmd.SetArgs(args)

	// Capture output
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	// Execute command
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Init command failed: %v\nStderr: %s\nStdout: %s", err, errBuf.String(), outBuf.String())
	}

	// Verify output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file was not created: %s", outputPath)
	}

	// Compare generated output with expected
	expectedPath := filepath.Join(testCaseDir, "expected.yaml")
	compareInitFiles(t, outputPath, expectedPath)
}

// compareInitFiles compares two YAML files and reports differences
func compareInitFiles(t *testing.T, generatedPath, expectedPath string) {
	t.Helper()

	generated, err := os.ReadFile(generatedPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	expected, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read expected file: %v", err)
	}

	generatedStr := normalizeInitOutput(string(generated))
	expectedStr := normalizeInitOutput(string(expected))

	if generatedStr != expectedStr {
		t.Errorf("Output mismatch:\n\nExpected:\n%s\n\nGot:\n%s\n\nFirst difference:\n%s",
			expectedStr,
			generatedStr,
			findFirstDifference(expectedStr, generatedStr),
		)
	}
}

// normalizeInitOutput normalizes whitespace and line endings for comparison
func normalizeInitOutput(s string) string {
	// Normalize line endings
	s = strings.ReplaceAll(s, "\r\n", "\n")

	// Trim trailing whitespace from each line
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

// TestInitCommand_ValidationError tests error handling for invalid YAML
func TestInitCommand_ValidationError(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "invalid.yaml")
	outputPath := filepath.Join(tmpDir, "output.yaml")

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

// TestInitCommand_MissingFlags tests error handling when required flags are missing
func TestInitCommand_MissingFlags(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "values.yaml")
	outputPath := filepath.Join(tmpDir, "output.yaml")

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

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when apiVersion is missing, but command succeeded")
	}

	// Test missing kind
	cmd2 := newInitCommand()
	cmd2.SetArgs([]string{
		inputPath,
		"--api-version", "test.io/v1",
		"-o", outputPath,
	})

	err = cmd2.Execute()
	if err == nil {
		t.Error("Expected error when kind is missing, but command succeeded")
	}
}

// TestInitCommand_NonExistentInputFile tests error handling for explicitly specified non-existent files
func TestInitCommand_NonExistentInputFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.yaml")

	cmd := newInitCommand()
	cmd.SetArgs([]string{
		"nonexistent.yaml",
		"--api-version", "test.io/v1",
		"--kind", "Test",
		"-o", outputPath,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error for non-existent input file, but command succeeded")
	}

	if !strings.Contains(err.Error(), "failed to read") {
		t.Errorf("Expected 'failed to read' error, got: %v", err)
	}
}

// TestInitCommand_OnlyOneKRMField tests error handling when only apiVersion or kind is present
func TestInitCommand_OnlyOneKRMField(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.yaml")

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

	err = cmd2.Execute()
	if err == nil {
		t.Error("Expected error when file has kind but missing apiVersion")
	} else if !strings.Contains(err.Error(), "apiVersion") {
		t.Errorf("Expected error about missing apiVersion, got: %v", err)
	}
}
