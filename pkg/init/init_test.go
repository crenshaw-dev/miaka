package init

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConvertToKRM_BasicConversion(t *testing.T) {
	// Create temp directory for test files
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "values.yaml")
	outputFile := filepath.Join(tmpDir, "example.values.yaml")

	// Create input values.yaml
	inputContent := `replicaCount: 3
image: nginx:latest
service:
  port: 80
  type: ClusterIP
`
	if err := os.WriteFile(inputFile, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	// Run conversion
	err := ConvertToKRM(inputFile, outputFile, "myapp.io/v1", "MyApp")
	if err != nil {
		t.Fatalf("ConvertToKRM failed: %v", err)
	}

	// Read output file
	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(output)

	// Verify apiVersion and kind are present
	if !strings.Contains(outputStr, "apiVersion: myapp.io/v1") {
		t.Error("Output does not contain apiVersion")
	}
	if !strings.Contains(outputStr, "kind: MyApp") {
		t.Error("Output does not contain kind")
	}

	// Verify original fields are preserved
	if !strings.Contains(outputStr, "replicaCount: 3") {
		t.Error("Output does not contain original replicaCount field")
	}
	if !strings.Contains(outputStr, "image: nginx:latest") {
		t.Error("Output does not contain original image field")
	}
}

func TestConvertToKRM_FieldOrdering(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "values.yaml")
	outputFile := filepath.Join(tmpDir, "example.values.yaml")

	inputContent := `foo: bar
baz: qux
`
	if err := os.WriteFile(inputFile, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err := ConvertToKRM(inputFile, outputFile, "test.io/v1", "Test")
	if err != nil {
		t.Fatalf("ConvertToKRM failed: %v", err)
	}

	// Parse output to verify field order
	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var node yaml.Node
	if err := yaml.Unmarshal(output, &node); err != nil {
		t.Fatalf("Failed to parse output YAML: %v", err)
	}

	// Get the mapping node
	contentNode := node.Content[0]
	if contentNode.Kind != yaml.MappingNode {
		t.Fatal("Expected mapping node")
	}

	// Check field order: apiVersion, kind, then others
	expectedOrder := []string{"apiVersion", "kind", "foo", "baz"}
	actualOrder := []string{}
	for i := 0; i < len(contentNode.Content); i += 2 {
		actualOrder = append(actualOrder, contentNode.Content[i].Value)
	}

	if len(actualOrder) < 4 {
		t.Fatalf("Expected at least 4 fields, got %d", len(actualOrder))
	}

	for i, expected := range expectedOrder {
		if actualOrder[i] != expected {
			t.Errorf("Field order mismatch at position %d: expected %s, got %s", i, expected, actualOrder[i])
		}
	}
}

func TestConvertToKRM_CommentPreservation(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "values.yaml")
	outputFile := filepath.Join(tmpDir, "example.values.yaml")

	// Input with comments
	inputContent := `# Replica count
replicaCount: 3
# Docker image
image: nginx:latest
`
	if err := os.WriteFile(inputFile, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err := ConvertToKRM(inputFile, outputFile, "test.io/v1", "Test")
	if err != nil {
		t.Fatalf("ConvertToKRM failed: %v", err)
	}

	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(output)

	// Verify comments are preserved
	if !strings.Contains(outputStr, "# Replica count") {
		t.Error("Comment for replicaCount was not preserved")
	}
	if !strings.Contains(outputStr, "# Docker image") {
		t.Error("Comment for image was not preserved")
	}
}

func TestConvertToKRM_NestedObjects(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "values.yaml")
	outputFile := filepath.Join(tmpDir, "example.values.yaml")

	inputContent := `service:
  port: 80
  type: ClusterIP
  annotations:
    foo: bar
`
	if err := os.WriteFile(inputFile, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err := ConvertToKRM(inputFile, outputFile, "test.io/v1", "Test")
	if err != nil {
		t.Fatalf("ConvertToKRM failed: %v", err)
	}

	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(output)

	// Verify nested structure is preserved
	if !strings.Contains(outputStr, "service:") {
		t.Error("Nested service object not preserved")
	}
	if !strings.Contains(outputStr, "port: 80") {
		t.Error("Nested service.port not preserved")
	}
	if !strings.Contains(outputStr, "annotations:") {
		t.Error("Deeply nested annotations not preserved")
	}
}

func TestConvertToKRM_Arrays(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "values.yaml")
	outputFile := filepath.Join(tmpDir, "example.values.yaml")

	inputContent := `env:
- name: FOO
  value: bar
- name: BAZ
  value: qux
`
	if err := os.WriteFile(inputFile, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err := ConvertToKRM(inputFile, outputFile, "test.io/v1", "Test")
	if err != nil {
		t.Fatalf("ConvertToKRM failed: %v", err)
	}

	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(output)

	// Verify array is preserved
	if !strings.Contains(outputStr, "env:") {
		t.Error("Array field not preserved")
	}
	if !strings.Contains(outputStr, "name: FOO") {
		t.Error("Array item not preserved")
	}
	if !strings.Contains(outputStr, "value: bar") {
		t.Error("Array item value not preserved")
	}
}

func TestConvertToKRM_AlreadyHasAPIVersion(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "values.yaml")
	outputFile := filepath.Join(tmpDir, "example.values.yaml")

	// Input already has both apiVersion and kind - should succeed
	inputContent := `apiVersion: v1
kind: ConfigMap
data:
  foo: bar
`
	if err := os.WriteFile(inputFile, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	// Should succeed without providing flags (uses existing values)
	err := ConvertToKRM(inputFile, outputFile, "", "")
	if err != nil {
		t.Errorf("Expected success when file has both apiVersion and kind, got error: %v", err)
	}
}

func TestConvertToKRM_AlreadyHasKind(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "values.yaml")
	outputFile := filepath.Join(tmpDir, "example.values.yaml")

	// Input already has kind but missing apiVersion
	inputContent := `kind: ConfigMap
data:
  foo: bar
`
	if err := os.WriteFile(inputFile, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	// Should return error asking for apiVersion flag
	err := ConvertToKRM(inputFile, outputFile, "", "")
	if err == nil {
		t.Error("Expected error when file has kind but missing apiVersion, got nil")
	}
	if !strings.Contains(err.Error(), "missing apiVersion") {
		t.Errorf("Expected error about missing apiVersion, got: %v", err)
	}
}

func TestConvertToKRM_InvalidInputFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "example.values.yaml")

	// Non-existent input file
	err := ConvertToKRM("nonexistent.yaml", outputFile, "test.io/v1", "Test")
	if err == nil {
		t.Error("Expected error for non-existent input file, got nil")
	}
}

func TestConvertToKRM_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "values.yaml")
	outputFile := filepath.Join(tmpDir, "example.values.yaml")

	// Invalid YAML
	inputContent := `this is not: valid: yaml: content`
	if err := os.WriteFile(inputFile, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err := ConvertToKRM(inputFile, outputFile, "test.io/v1", "Test")
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestConvertToKRM_EmptyInputFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "example.values.yaml")

	// Test creating empty file with no input
	err := ConvertToKRM("", outputFile, "test.io/v1", "Test")
	if err != nil {
		t.Fatalf("ConvertToKRM failed: %v", err)
	}

	// Read output file
	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(output)

	// Verify apiVersion and kind are present
	if !strings.Contains(outputStr, "apiVersion: test.io/v1") {
		t.Error("Output does not contain apiVersion")
	}
	if !strings.Contains(outputStr, "kind: Test") {
		t.Error("Output does not contain kind")
	}

	// Verify no other fields
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines (apiVersion and kind), got %d", len(lines))
	}
}

func TestConvertToKRM_EmptyInputFileRequiresFlags(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "example.values.yaml")

	// Should fail without apiVersion
	err := ConvertToKRM("", outputFile, "", "Test")
	if err == nil {
		t.Error("Expected error when creating empty file without apiVersion, got nil")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected error about required flags, got: %v", err)
	}

	// Should fail without kind
	err = ConvertToKRM("", outputFile, "test.io/v1", "")
	if err == nil {
		t.Error("Expected error when creating empty file without kind, got nil")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected error about required flags, got: %v", err)
	}
}

func TestConvertToKRM_MissingFieldsRequireFlags(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "values.yaml")
	outputFile := filepath.Join(tmpDir, "example.values.yaml")

	// Create input without apiVersion or kind
	inputContent := `replicaCount: 3
image: nginx:latest
`
	if err := os.WriteFile(inputFile, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	// Should fail without providing flags
	err := ConvertToKRM(inputFile, outputFile, "", "")
	if err == nil {
		t.Error("Expected error when input doesn't have apiVersion/kind and flags aren't provided, got nil")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected error about required fields, got: %v", err)
	}

	// Should succeed when providing flags
	err = ConvertToKRM(inputFile, outputFile, "test.io/v1", "Test")
	if err != nil {
		t.Fatalf("ConvertToKRM failed: %v", err)
	}

	// Read output file
	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(output)

	// Verify apiVersion and kind are added
	if !strings.Contains(outputStr, "apiVersion: test.io/v1") {
		t.Error("Output does not contain apiVersion")
	}
	if !strings.Contains(outputStr, "kind: Test") {
		t.Error("Output does not contain kind")
	}
	// Verify original fields preserved
	if !strings.Contains(outputStr, "replicaCount: 3") {
		t.Error("Output does not contain original replicaCount")
	}
}

func TestCheckKRMFields_BothPresent(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.yaml")

	content := `apiVersion: v1
kind: ConfigMap
data:
  foo: bar
`
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	hasApiVersion, hasKind := CheckKRMFields(inputFile)
	if !hasApiVersion {
		t.Error("Expected hasApiVersion to be true")
	}
	if !hasKind {
		t.Error("Expected hasKind to be true")
	}
}

func TestCheckKRMFields_NeitherPresent(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.yaml")

	content := `replicaCount: 3
image: nginx:latest
`
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	hasApiVersion, hasKind := CheckKRMFields(inputFile)
	if hasApiVersion {
		t.Error("Expected hasApiVersion to be false")
	}
	if hasKind {
		t.Error("Expected hasKind to be false")
	}
}

func TestCheckKRMFields_OnlyApiVersion(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.yaml")

	content := `apiVersion: v1
data:
  foo: bar
`
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	hasApiVersion, hasKind := CheckKRMFields(inputFile)
	if !hasApiVersion {
		t.Error("Expected hasApiVersion to be true")
	}
	if hasKind {
		t.Error("Expected hasKind to be false")
	}
}

func TestCheckKRMFields_OnlyKind(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.yaml")

	content := `kind: ConfigMap
data:
  foo: bar
`
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	hasApiVersion, hasKind := CheckKRMFields(inputFile)
	if hasApiVersion {
		t.Error("Expected hasApiVersion to be false")
	}
	if !hasKind {
		t.Error("Expected hasKind to be true")
	}
}

func TestCheckKRMFields_NonExistentFile(t *testing.T) {
	hasApiVersion, hasKind := CheckKRMFields("nonexistent.yaml")
	if hasApiVersion {
		t.Error("Expected hasApiVersion to be false for non-existent file")
	}
	if hasKind {
		t.Error("Expected hasKind to be false for non-existent file")
	}
}

func TestCheckKRMFields_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.yaml")

	content := `this is not: valid: yaml: content`
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	hasApiVersion, hasKind := CheckKRMFields(inputFile)
	if hasApiVersion {
		t.Error("Expected hasApiVersion to be false for invalid YAML")
	}
	if hasKind {
		t.Error("Expected hasKind to be false for invalid YAML")
	}
}
