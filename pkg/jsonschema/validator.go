package jsonschema

import (
	"bytes"
	"fmt"
	"os"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"sigs.k8s.io/yaml"
)

// ValidateYAML validates a YAML file against a JSON Schema file
// Uses the same validation method as Helm for maximum compatibility
func ValidateYAML(yamlPath, schemaPath string) error {
	// Read and parse YAML file
	yamlBytes, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	// Unmarshal YAML to map (same as Helm does with values)
	var values map[string]interface{}
	if err := yaml.Unmarshal(yamlBytes, &values); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Read JSON Schema
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Validate using Helm's approach
	return validateAgainstSchema(values, schemaBytes)
}

// validateAgainstSchema checks that values conform to the JSON Schema
// This follows Helm's exact validation pattern from:
// https://github.com/helm/helm/blob/main/pkg/chart/common/util/jsonschema.go
func validateAgainstSchema(values map[string]interface{}, schemaJSON []byte) error {
	// Unmarshal schema (Helm uses UnmarshalJSON which leverages UseNumber)
	schema, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaJSON))
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON Schema: %w", err)
	}

	// Create compiler (following Helm's pattern)
	compiler := jsonschema.NewCompiler()
	
	// Add schema resource
	err = compiler.AddResource("file:///values.schema.json", schema)
	if err != nil {
		return fmt.Errorf("failed to add schema resource: %w", err)
	}

	// Compile schema
	validator, err := compiler.Compile("file:///values.schema.json")
	if err != nil {
		return fmt.Errorf("failed to compile schema: %w", err)
	}

	// Validate values
	err = validator.Validate(values)
	if err != nil {
		return fmt.Errorf("JSON Schema validation failed: %w", err)
	}

	return nil
}

