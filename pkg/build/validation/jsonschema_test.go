package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateYAML_ValidData(t *testing.T) {
	tmpDir := t.TempDir()

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
	err := os.WriteFile(schemaPath, []byte(schemaContent), 0644)
	require.NoError(t, err, "Failed to write schema")

	// Create valid YAML
	yamlPath := filepath.Join(tmpDir, "valid.yaml")
	yamlContent := `name: test
count: 5
`
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err, "Failed to write YAML")

	// Test validation
	err = ValidateYAML(yamlPath, schemaPath)
	assert.NoError(t, err, "ValidateYAML() expected no error for valid data")
}

func TestValidateYAML_InvalidData_MissingRequired(t *testing.T) {
	tmpDir := t.TempDir()

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
	err := os.WriteFile(schemaPath, []byte(schemaContent), 0644)
	require.NoError(t, err, "Failed to write schema")

	// Create YAML missing required field
	yamlPath := filepath.Join(tmpDir, "invalid.yaml")
	yamlContent := `count: 5
`
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err, "Failed to write YAML")

	// Test validation
	err = ValidateYAML(yamlPath, schemaPath)
	assert.Error(t, err, "ValidateYAML() expected error for missing required field")
}

func TestValidateYAML_InvalidData_WrongType(t *testing.T) {
	tmpDir := t.TempDir()

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
	err := os.WriteFile(schemaPath, []byte(schemaContent), 0644)
	require.NoError(t, err, "Failed to write schema")

	// Create YAML with wrong type
	yamlPath := filepath.Join(tmpDir, "invalid.yaml")
	yamlContent := `count: "not a number"
`
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err, "Failed to write YAML")

	// Test validation
	err = ValidateYAML(yamlPath, schemaPath)
	assert.Error(t, err, "ValidateYAML() expected error for wrong type")
}

func TestValidateYAML_InvalidData_ViolatesMinimum(t *testing.T) {
	tmpDir := t.TempDir()

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
	err := os.WriteFile(schemaPath, []byte(schemaContent), 0644)
	require.NoError(t, err, "Failed to write schema")

	// Create YAML that violates minimum
	yamlPath := filepath.Join(tmpDir, "invalid.yaml")
	yamlContent := `count: 0
`
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err, "Failed to write YAML")

	// Test validation
	err = ValidateYAML(yamlPath, schemaPath)
	assert.Error(t, err, "ValidateYAML() expected error for violating minimum")
}

func TestValidateYAML_MissingYAMLFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create schema but no YAML file
	schemaPath := filepath.Join(tmpDir, "schema.json")
	schemaContent := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object"
}`
	err := os.WriteFile(schemaPath, []byte(schemaContent), 0644)
	require.NoError(t, err, "Failed to write schema")

	yamlPath := filepath.Join(tmpDir, "nonexistent.yaml")

	// Test validation
	err = ValidateYAML(yamlPath, schemaPath)
	assert.Error(t, err, "ValidateYAML() expected error for missing YAML file")
}

func TestValidateYAML_MissingSchemaFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create YAML but no schema file
	yamlPath := filepath.Join(tmpDir, "test.yaml")
	yamlContent := `name: test
`
	err := os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err, "Failed to write YAML")

	schemaPath := filepath.Join(tmpDir, "nonexistent.json")

	// Test validation
	err = ValidateYAML(yamlPath, schemaPath)
	assert.Error(t, err, "ValidateYAML() expected error for missing schema file")
}

func TestValidateYAML_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// Create schema
	schemaPath := filepath.Join(tmpDir, "schema.json")
	schemaContent := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object"
}`
	err := os.WriteFile(schemaPath, []byte(schemaContent), 0644)
	require.NoError(t, err, "Failed to write schema")

	// Create invalid YAML
	yamlPath := filepath.Join(tmpDir, "invalid.yaml")
	yamlContent := `{
  this is not: valid: yaml:
`
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err, "Failed to write YAML")

	// Test validation
	err = ValidateYAML(yamlPath, schemaPath)
	assert.Error(t, err, "ValidateYAML() expected error for invalid YAML")
}

func TestValidateYAML_InvalidSchema(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid JSON schema
	schemaPath := filepath.Join(tmpDir, "schema.json")
	schemaContent := `{
  this is not valid json
}`
	err := os.WriteFile(schemaPath, []byte(schemaContent), 0644)
	require.NoError(t, err, "Failed to write schema")

	// Create valid YAML
	yamlPath := filepath.Join(tmpDir, "test.yaml")
	yamlContent := `name: test
`
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err, "Failed to write YAML")

	// Test validation
	err = ValidateYAML(yamlPath, schemaPath)
	assert.Error(t, err, "ValidateYAML() expected error for invalid schema")
}
