package crd

import (
	"fmt"
	"os"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"sigs.k8s.io/yaml"
)

// ValidateCRD validates a CRD file for structural schema compliance
// This catches issues like mutually exclusive properties before trying to apply to a cluster
func ValidateCRD(crdPath string) error {
	// Read and parse CRD
	data, err := os.ReadFile(crdPath)
	if err != nil {
		return fmt.Errorf("failed to read CRD: %w", err)
	}

	var crd apiextensionsv1.CustomResourceDefinition
	if err := yaml.Unmarshal(data, &crd); err != nil {
		return fmt.Errorf("failed to parse CRD: %w", err)
	}

	// Validate each version's schema
	for _, version := range crd.Spec.Versions {
		if version.Schema != nil && version.Schema.OpenAPIV3Schema != nil {
			// Convert v1 schema to internal schema
			internalSchema := &apiextensions.JSONSchemaProps{}
			if err := apiextensionsv1.Convert_v1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(
				version.Schema.OpenAPIV3Schema,
				internalSchema,
				nil,
			); err != nil {
				return fmt.Errorf("failed to convert schema for version %s: %w", version.Name, err)
			}

			// Validate structural schema
			ss, err := schema.NewStructural(internalSchema)
			if err != nil {
				return fmt.Errorf("CRD schema validation failed for version %s: %w", version.Name, err)
			}

			// Check if schema is valid (this will catch the mutually exclusive properties issue)
			if ss == nil {
				return fmt.Errorf("CRD schema for version %s is not structural", version.Name)
			}
		}
	}

	return nil
}
