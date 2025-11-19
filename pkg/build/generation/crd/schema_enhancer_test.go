package crd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

func TestAddStrictValidation_Success(t *testing.T) {
	tmpDir := t.TempDir()
	crdPath := filepath.Join(tmpDir, "crd.yaml")

	// Create a CRD with object types that need additionalProperties: false
	crd := &apiextensionsv1.CustomResourceDefinition{
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "example.com",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:   "Example",
				Plural: "examples",
			},
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name: "v1",
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"apiVersion": {Type: "string"},
								"kind":       {Type: "string"},
								"metadata": {
									Type: "object",
								},
								"spec": {
									Type: "object",
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										"config": {
											Type: "object", // This should get additionalProperties: false
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	data, err := yaml.Marshal(crd)
	require.NoError(t, err)
	err = os.WriteFile(crdPath, data, 0644)
	require.NoError(t, err)

	// Add strict validation
	err = AddStrictValidation(crdPath)
	require.NoError(t, err)

	// Read back the CRD
	modifiedData, err := os.ReadFile(crdPath)
	require.NoError(t, err)

	var modifiedCRD apiextensionsv1.CustomResourceDefinition
	err = yaml.Unmarshal(modifiedData, &modifiedCRD)
	require.NoError(t, err)

	// Verify metadata was removed from properties
	schema := modifiedCRD.Spec.Versions[0].Schema.OpenAPIV3Schema
	_, hasMetadata := schema.Properties["metadata"]
	assert.False(t, hasMetadata, "metadata should be removed from properties")

	// Verify config object has additionalProperties: false
	specProps := schema.Properties["spec"]
	configProp := specProps.Properties["config"]
	require.NotNil(t, configProp.AdditionalProperties, "config should have additionalProperties set")
	assert.False(t, configProp.AdditionalProperties.Allows, "config should have additionalProperties: false")
}

func TestAddStrictValidation_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	crdPath := filepath.Join(tmpDir, "nonexistent.yaml")

	err := AddStrictValidation(crdPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read CRD")
}

func TestAddStrictValidation_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	crdPath := filepath.Join(tmpDir, "invalid.yaml")

	err := os.WriteFile(crdPath, []byte("invalid: yaml: [[["), 0644)
	require.NoError(t, err)

	err = AddStrictValidation(crdPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse CRD")
}

func TestAddStrictValidation_MultipleVersions(t *testing.T) {
	tmpDir := t.TempDir()
	crdPath := filepath.Join(tmpDir, "crd.yaml")

	// Create a CRD with multiple versions
	crd := &apiextensionsv1.CustomResourceDefinition{
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "example.com",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:   "Example",
				Plural: "examples",
			},
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name: "v1alpha1",
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"spec": {
									Type: "object",
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										"data": {
											Type: "object",
										},
									},
								},
							},
						},
					},
				},
				{
					Name: "v1",
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"spec": {
									Type: "object",
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										"config": {
											Type: "object",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	data, err := yaml.Marshal(crd)
	require.NoError(t, err)
	err = os.WriteFile(crdPath, data, 0644)
	require.NoError(t, err)

	// Add strict validation
	err = AddStrictValidation(crdPath)
	require.NoError(t, err)

	// Read back and verify both versions were processed
	modifiedData, err := os.ReadFile(crdPath)
	require.NoError(t, err)

	var modifiedCRD apiextensionsv1.CustomResourceDefinition
	err = yaml.Unmarshal(modifiedData, &modifiedCRD)
	require.NoError(t, err)

	// Check v1alpha1
	v1alpha1Schema := modifiedCRD.Spec.Versions[0].Schema.OpenAPIV3Schema
	v1alpha1SpecProps := v1alpha1Schema.Properties["spec"]
	v1alpha1DataProp := v1alpha1SpecProps.Properties["data"]
	require.NotNil(t, v1alpha1DataProp.AdditionalProperties)
	assert.False(t, v1alpha1DataProp.AdditionalProperties.Allows)

	// Check v1
	v1Schema := modifiedCRD.Spec.Versions[1].Schema.OpenAPIV3Schema
	v1SpecProps := v1Schema.Properties["spec"]
	v1ConfigProp := v1SpecProps.Properties["config"]
	require.NotNil(t, v1ConfigProp.AdditionalProperties)
	assert.False(t, v1ConfigProp.AdditionalProperties.Allows)
}

func TestAddAdditionalPropertiesFalse_NestedStructures(t *testing.T) {
	// Test with nested objects in arrays
	schema := &apiextensionsv1.JSONSchemaProps{
		Type: "object",
		Properties: map[string]apiextensionsv1.JSONSchemaProps{
			"items": {
				Type: "array",
				Items: &apiextensionsv1.JSONSchemaPropsOrArray{
					Schema: &apiextensionsv1.JSONSchemaProps{
						Type: "object",
						Properties: map[string]apiextensionsv1.JSONSchemaProps{
							"nested": {
								Type: "object",
							},
						},
					},
				},
			},
		},
	}

	addAdditionalPropertiesFalse(schema)

	// Verify nested object in array has additionalProperties: false
	itemsSchema := schema.Properties["items"].Items.Schema
	nestedProp := itemsSchema.Properties["nested"]
	require.NotNil(t, nestedProp.AdditionalProperties)
	assert.False(t, nestedProp.AdditionalProperties.Allows)
}

func TestAddAdditionalPropertiesFalse_AllOfAnyOfOneOf(t *testing.T) {
	schema := &apiextensionsv1.JSONSchemaProps{
		Type: "object",
		AllOf: []apiextensionsv1.JSONSchemaProps{
			{
				Type: "object",
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"field1": {Type: "object"},
				},
			},
		},
		AnyOf: []apiextensionsv1.JSONSchemaProps{
			{
				Type: "object",
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"field2": {Type: "object"},
				},
			},
		},
		OneOf: []apiextensionsv1.JSONSchemaProps{
			{
				Type: "object",
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"field3": {Type: "object"},
				},
			},
		},
	}

	addAdditionalPropertiesFalse(schema)

	// Verify allOf
	field1 := schema.AllOf[0].Properties["field1"]
	require.NotNil(t, field1.AdditionalProperties)
	assert.False(t, field1.AdditionalProperties.Allows)

	// Verify anyOf
	field2 := schema.AnyOf[0].Properties["field2"]
	require.NotNil(t, field2.AdditionalProperties)
	assert.False(t, field2.AdditionalProperties.Allows)

	// Verify oneOf
	field3 := schema.OneOf[0].Properties["field3"]
	require.NotNil(t, field3.AdditionalProperties)
	assert.False(t, field3.AdditionalProperties.Allows)
}

func TestAddAdditionalPropertiesFalse_ObjectWithProperties(t *testing.T) {
	// Objects with properties should NOT get additionalProperties: false
	// (they're mutually exclusive in structural schemas)
	schema := &apiextensionsv1.JSONSchemaProps{
		Type: "object",
		Properties: map[string]apiextensionsv1.JSONSchemaProps{
			"field": {Type: "string"},
		},
	}

	addAdditionalPropertiesFalse(schema)

	// Should NOT have additionalProperties set because it has properties
	assert.Nil(t, schema.AdditionalProperties, "object with properties should not have additionalProperties")
}

func TestAddAdditionalPropertiesFalse_ObjectWithExistingAdditionalProperties(t *testing.T) {
	// Objects with existing additionalProperties should not be modified
	schema := &apiextensionsv1.JSONSchemaProps{
		Type: "object",
		AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{
			Allows: true,
			Schema: &apiextensionsv1.JSONSchemaProps{
				Type: "string",
			},
		},
	}

	addAdditionalPropertiesFalse(schema)

	// Should keep existing additionalProperties
	require.NotNil(t, schema.AdditionalProperties)
	assert.True(t, schema.AdditionalProperties.Allows, "existing additionalProperties should not be changed")
}

func TestAddAdditionalPropertiesFalse_NilSchema(t *testing.T) {
	// Should not panic on nil schema
	addAdditionalPropertiesFalse(nil)
	// Just testing that it doesn't crash
}

func TestAddAdditionalPropertiesFalse_WithAdditionalPropertiesSchema(t *testing.T) {
	// Test recursion into additionalProperties.Schema
	schema := &apiextensionsv1.JSONSchemaProps{
		Type: "object",
		AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{
			Allows: true,
			Schema: &apiextensionsv1.JSONSchemaProps{
				Type: "object",
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"nested": {
						Type: "object",
					},
				},
			},
		},
	}

	addAdditionalPropertiesFalse(schema)

	// Verify nested object in additionalProperties.Schema gets processed
	nestedProp := schema.AdditionalProperties.Schema.Properties["nested"]
	require.NotNil(t, nestedProp.AdditionalProperties)
	assert.False(t, nestedProp.AdditionalProperties.Allows)
}
