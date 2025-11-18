package crdgen

import (
	"fmt"
	"os"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

// AddStrictValidation adds additionalProperties: false to all object schemas in a CRD
// This ensures strict validation that rejects unknown fields
func AddStrictValidation(crdPath string) error {
	// Read and parse CRD
	data, err := os.ReadFile(crdPath)
	if err != nil {
		return fmt.Errorf("failed to read CRD: %w", err)
	}

	var crd apiextensionsv1.CustomResourceDefinition
	if err := yaml.Unmarshal(data, &crd); err != nil {
		return fmt.Errorf("failed to parse CRD: %w", err)
	}

	// Process all versions
	for i := range crd.Spec.Versions {
		if crd.Spec.Versions[i].Schema != nil && crd.Spec.Versions[i].Schema.OpenAPIV3Schema != nil {
			schema := crd.Spec.Versions[i].Schema.OpenAPIV3Schema

			// Remove metadata from properties if present (Kubernetes manages this)
			if schema.Properties != nil {
				delete(schema.Properties, "metadata")
			}

			addAdditionalPropertiesFalse(schema)
		}
	}

	// Write back
	output, err := yaml.Marshal(&crd)
	if err != nil {
		return fmt.Errorf("failed to marshal CRD: %w", err)
	}

	if err := os.WriteFile(crdPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write CRD: %w", err)
	}

	return nil
}

// addAdditionalPropertiesFalse recursively sets additionalProperties to false for all object types
func addAdditionalPropertiesFalse(schema *apiextensionsv1.JSONSchemaProps) {
	if schema == nil {
		return
	}

	// If this is an object type, add additionalProperties: false
	// BUT only if:
	// 1. It doesn't already have properties defined (properties and additionalProperties are mutually exclusive in structural schemas)
	// 2. It doesn't already have additionalProperties set
	if schema.Type == "object" && len(schema.Properties) == 0 && schema.AdditionalProperties == nil {
		schema.AdditionalProperties = &apiextensionsv1.JSONSchemaPropsOrBool{
			Allows: false,
			Schema: nil,
		}
	}

	// Recurse into properties
	for key := range schema.Properties {
		prop := schema.Properties[key]
		addAdditionalPropertiesFalse(&prop)
		schema.Properties[key] = prop
	}

	// Recurse into array items
	if schema.Items != nil && schema.Items.Schema != nil {
		addAdditionalPropertiesFalse(schema.Items.Schema)
	}

	// Recurse into additionalProperties if it has a schema
	if schema.AdditionalProperties != nil && schema.AdditionalProperties.Schema != nil {
		addAdditionalPropertiesFalse(schema.AdditionalProperties.Schema)
	}

	// Recurse into allOf, anyOf, oneOf
	for i := range schema.AllOf {
		addAdditionalPropertiesFalse(&schema.AllOf[i])
	}
	for i := range schema.AnyOf {
		addAdditionalPropertiesFalse(&schema.AnyOf[i])
	}
	for i := range schema.OneOf {
		addAdditionalPropertiesFalse(&schema.OneOf[i])
	}
}
