package validator

import (
	"fmt"
	"os"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

// ValidateAgainstCRD validates a resource YAML file against a CRD
// Returns an error if validation fails
func ValidateAgainstCRD(crdPath, resourcePath string) error {
	// Load and unmarshal CRD
	crdData, err := os.ReadFile(crdPath)
	if err != nil {
		return fmt.Errorf("failed to read CRD file: %w", err)
	}

	crd := &apiextensionsv1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(crdData, crd); err != nil {
		return fmt.Errorf("failed to unmarshal CRD: %w", err)
	}

	// Load and unmarshal resource
	resourceData, err := os.ReadFile(resourcePath)
	if err != nil {
		return fmt.Errorf("failed to read resource file: %w", err)
	}

	resource := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(resourceData, resource); err != nil {
		return fmt.Errorf("failed to unmarshal resource: %w", err)
	}

	// Find the schema for the resource's version
	var schema *apiextensionsv1.CustomResourceValidation
	resourceVersion := resource.GetAPIVersion()

	for _, version := range crd.Spec.Versions {
		// Check if this version matches the resource
		expectedAPIVersion := crd.Spec.Group + "/" + version.Name
		if expectedAPIVersion == resourceVersion {
			schema = version.Schema
			break
		}
	}

	if schema == nil || schema.OpenAPIV3Schema == nil {
		return fmt.Errorf("no schema found for version %s in CRD", resourceVersion)
	}

	// Convert the v1 schema to internal schema for validation
	internalSchema := &apiextensions.JSONSchemaProps{}
	if err := apiextensionsv1.Convert_v1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(
		schema.OpenAPIV3Schema,
		internalSchema,
		nil,
	); err != nil {
		return fmt.Errorf("failed to convert schema: %w", err)
	}

	// Create schema validator
	schemaValidator, _, err := validation.NewSchemaValidator(internalSchema)
	if err != nil {
		return fmt.Errorf("failed to create schema validator: %w", err)
	}

	// Validate the resource
	result := schemaValidator.Validate(resource.Object)
	if len(result.Errors) > 0 {
		return fmt.Errorf("resource validation failed:\n%v", result.Errors)
	}

	// Note: We ignore warnings for now, only fail on errors
	return nil
}
