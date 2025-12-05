// Package jsonschema generates JSON Schema files for Helm chart validation.
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

	// Add title and additionalProperties to all objects (including root)
	// First handle root level
	if typeVal, ok := schema["type"].(string); ok && typeVal == "object" {
		if _, exists := schema["additionalProperties"]; !exists {
			schema["additionalProperties"] = false
		}
	}

	// Then enhance nested properties
	enhanceSchemaForHelm(schema, "")

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

// enhanceSchemaForHelm adds title and additionalProperties fields to match helm-schema output
func enhanceSchemaForHelm(obj interface{}, propertyName string) {
	switch v := obj.(type) {
	case map[string]interface{}:
		setAdditionalPropertiesIfNeeded(v)
		setTitleIfNeeded(v, propertyName)
		recurseIntoNestedSchemas(v)
	case []interface{}:
		for _, item := range v {
			enhanceSchemaForHelm(item, "")
		}
	}
}

// setAdditionalPropertiesIfNeeded sets additionalProperties: false for objects
func setAdditionalPropertiesIfNeeded(v map[string]interface{}) {
	if typeVal, ok := v["type"].(string); ok && typeVal == "object" {
		if _, exists := v["additionalProperties"]; !exists {
			v["additionalProperties"] = false
		}
	}
}

// setTitleIfNeeded adds title field if property name is provided
func setTitleIfNeeded(v map[string]interface{}, propertyName string) {
	if propertyName != "" {
		if _, exists := v["title"]; !exists {
			v["title"] = propertyName
		}
	}
}

// recurseIntoNestedSchemas recursively processes nested schemas
func recurseIntoNestedSchemas(v map[string]interface{}) {
	// Recurse into properties
	if properties, ok := v["properties"].(map[string]interface{}); ok {
		for propName, propValue := range properties {
			enhanceSchemaForHelm(propValue, propName)
		}
	}

	// Recurse into items (for arrays)
	if items, ok := v["items"]; ok {
		enhanceSchemaForHelm(items, "")
	}

	// Recurse into other nested objects
	for key, value := range v {
		if key != "properties" && key != "items" {
			switch value.(type) {
			case map[string]interface{}, []interface{}:
				enhanceSchemaForHelm(value, "")
			}
		}
	}
}
