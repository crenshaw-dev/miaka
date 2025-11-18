package jsonschema

import (
	"encoding/json"
	"fmt"
	"os"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

// GenerateFromCRD extracts the OpenAPI v3 schema from a CRD and converts it to JSON Schema
func GenerateFromCRD(crdPath, outputPath string) error {
	// Read CRD file
	crdBytes, err := os.ReadFile(crdPath)
	if err != nil {
		return fmt.Errorf("failed to read CRD file: %w", err)
	}

	// Parse CRD
	var crd apiextensionsv1.CustomResourceDefinition
	if err := yaml.Unmarshal(crdBytes, &crd); err != nil {
		return fmt.Errorf("failed to parse CRD YAML: %w", err)
	}

	// Find the first version with a schema
	var schema *apiextensionsv1.JSONSchemaProps
	for _, version := range crd.Spec.Versions {
		if version.Schema != nil && version.Schema.OpenAPIV3Schema != nil {
			schema = version.Schema.OpenAPIV3Schema
			break
		}
	}

	if schema == nil {
		return fmt.Errorf("no schema found in CRD")
	}

	// Convert to JSON Schema format
	jsonSchema, err := convertToJSONSchema(schema)
	if err != nil {
		return fmt.Errorf("failed to convert to JSON Schema: %w", err)
	}

	// Write JSON Schema file
	jsonBytes, err := json.MarshalIndent(jsonSchema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON Schema: %w", err)
	}

	if err := os.WriteFile(outputPath, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write JSON Schema file: %w", err)
	}

	return nil
}

// convertToJSONSchema converts an OpenAPI v3 schema to JSON Schema format
func convertToJSONSchema(openAPISchema *apiextensionsv1.JSONSchemaProps) (map[string]interface{}, error) {
	// Marshal to JSON first (this preserves all fields)
	openAPIBytes, err := json.Marshal(openAPISchema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OpenAPI schema: %w", err)
	}

	// Unmarshal into generic map
	var schema map[string]interface{}
	if err := json.Unmarshal(openAPIBytes, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	// Add JSON Schema metadata
	schema["$schema"] = "http://json-schema.org/draft-07/schema#"

	// Remove metadata field from properties if it exists
	if properties, ok := schema["properties"].(map[string]interface{}); ok {
		delete(properties, "metadata")
	}

	// Remove Kubernetes-specific extensions if present
	removeKubernetesExtensions(schema)

	return schema, nil
}

// removeKubernetesExtensions recursively removes x-kubernetes-* fields
func removeKubernetesExtensions(obj interface{}) {
	switch v := obj.(type) {
	case map[string]interface{}:
		// Remove x-kubernetes-* keys
		for key := range v {
			if len(key) > 13 && key[:13] == "x-kubernetes-" {
				delete(v, key)
			}
		}
		// Recurse into nested objects
		for _, value := range v {
			removeKubernetesExtensions(value)
		}
	case []interface{}:
		// Recurse into arrays
		for _, item := range v {
			removeKubernetesExtensions(item)
		}
	}
}
