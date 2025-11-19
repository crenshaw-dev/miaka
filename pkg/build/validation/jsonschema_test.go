package validation

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateYAML_ValidData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jsonschema-test-*")
	require.NoError(t, err, "Failed to create temp dir: %v")
	defer os.RemoveAll(tmpDir)

	// Create a simple JSON Schema
	schemaPath := filepath.Join(tmpDir, "schema.json")
	schemaContent := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "name": {
      "type": "string"
    },
    "count": {
      "type": "integer",
      "minimum": 0
    }
  },
  "required": ["name"]
}`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema: %v", err)
	}

	// Create valid YAML
	yamlPath := filepath.Join(tmpDir, "valid.yaml")
	yamlContent := `name: test
count: 5
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML: %v", err)
	}

	// Test validation
	err = ValidateYAML(yamlPath, schemaPath)
	assert.NoError(t, err, "ValidateYAML() expected no error for valid data")
}

func TestValidateYAML_InvalidData_MissingRequired(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jsonschema-test-*")
	require.NoError(t, err, "Failed to create temp dir: %v")
	defer os.RemoveAll(tmpDir)

	// Create JSON Schema with required field
	schemaPath := filepath.Join(tmpDir, "schema.json")
	schemaContent := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "name": {
      "type": "string"
    },
    "count": {
      "type": "integer"
    }
  },
  "required": ["name"]
}`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema: %v", err)
	}

	// Create YAML missing required field
	yamlPath := filepath.Join(tmpDir, "invalid.yaml")
	yamlContent := `count: 5
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML: %v", err)
	}

	// Test validation
	err = ValidateYAML(yamlPath, schemaPath)
	require.Error(t, err, "ValidateYAML() expected error for missing required field, got nil")
}

func TestValidateYAML_InvalidData_WrongType(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jsonschema-test-*")
	require.NoError(t, err, "Failed to create temp dir: %v")
	defer os.RemoveAll(tmpDir)

	// Create JSON Schema
	schemaPath := filepath.Join(tmpDir, "schema.json")
	schemaContent := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "count": {
      "type": "integer"
    }
  }
}`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema: %v", err)
	}

	// Create YAML with wrong type
	yamlPath := filepath.Join(tmpDir, "invalid.yaml")
	yamlContent := `count: "not a number"
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML: %v", err)
	}

	// Test validation
	err = ValidateYAML(yamlPath, schemaPath)
	require.Error(t, err, "ValidateYAML() expected error for wrong type, got nil")
}

func TestValidateYAML_InvalidData_ViolatesMinimum(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jsonschema-test-*")
	require.NoError(t, err, "Failed to create temp dir: %v")
	defer os.RemoveAll(tmpDir)

	// Create JSON Schema with minimum constraint
	schemaPath := filepath.Join(tmpDir, "schema.json")
	schemaContent := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "count": {
      "type": "integer",
      "minimum": 1
    }
  }
}`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema: %v", err)
	}

	// Create YAML that violates minimum
	yamlPath := filepath.Join(tmpDir, "invalid.yaml")
	yamlContent := `count: 0
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML: %v", err)
	}

	// Test validation
	err = ValidateYAML(yamlPath, schemaPath)
	require.Error(t, err, "ValidateYAML() expected error for violating minimum, got nil")
}

func TestValidateYAML_MissingYAMLFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jsonschema-test-*")
	require.NoError(t, err, "Failed to create temp dir: %v")
	defer os.RemoveAll(tmpDir)

	// Create schema but no YAML file
	schemaPath := filepath.Join(tmpDir, "schema.json")
	schemaContent := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object"
}`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema: %v", err)
	}

	yamlPath := filepath.Join(tmpDir, "nonexistent.yaml")

	// Test validation
	err = ValidateYAML(yamlPath, schemaPath)
	require.Error(t, err, "ValidateYAML() expected error for missing YAML file, got nil")
}

func TestValidateYAML_MissingSchemaFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jsonschema-test-*")
	require.NoError(t, err, "Failed to create temp dir: %v")
	defer os.RemoveAll(tmpDir)

	// Create YAML but no schema file
	yamlPath := filepath.Join(tmpDir, "test.yaml")
	yamlContent := `name: test
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML: %v", err)
	}

	schemaPath := filepath.Join(tmpDir, "nonexistent.json")

	// Test validation
	err = ValidateYAML(yamlPath, schemaPath)
	require.Error(t, err, "ValidateYAML() expected error for missing schema file, got nil")
}

func TestValidateYAML_InvalidYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jsonschema-test-*")
	require.NoError(t, err, "Failed to create temp dir: %v")
	defer os.RemoveAll(tmpDir)

	// Create schema
	schemaPath := filepath.Join(tmpDir, "schema.json")
	schemaContent := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object"
}`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema: %v", err)
	}

	// Create invalid YAML
	yamlPath := filepath.Join(tmpDir, "invalid.yaml")
	yamlContent := `{
  this is not: valid: yaml:
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML: %v", err)
	}

	// Test validation
	err = ValidateYAML(yamlPath, schemaPath)
	require.Error(t, err, "ValidateYAML() expected error for invalid YAML, got nil")
}

func TestValidateYAML_InvalidSchema(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jsonschema-test-*")
	require.NoError(t, err, "Failed to create temp dir: %v")
	defer os.RemoveAll(tmpDir)

	// Create invalid JSON schema
	schemaPath := filepath.Join(tmpDir, "schema.json")
	schemaContent := `{
  this is not valid json
}`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema: %v", err)
	}

	// Create valid YAML
	yamlPath := filepath.Join(tmpDir, "test.yaml")
	yamlContent := `name: test
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML: %v", err)
	}

	// Test validation
	err = ValidateYAML(yamlPath, schemaPath)
	require.Error(t, err, "ValidateYAML() expected error for invalid schema, got nil")
}
