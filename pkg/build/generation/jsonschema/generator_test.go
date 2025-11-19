package jsonschema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

func TestGenerateFromCRD_Success(t *testing.T) {
	tmpDir := t.TempDir()
	crdPath := filepath.Join(tmpDir, "crd.yaml")
	outputPath := filepath.Join(tmpDir, "schema.json")

	// Create a valid CRD
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
										"replicas": {Type: "integer"},
										"name":     {Type: "string"},
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

	// Generate JSON Schema
	err = GenerateFromCRD(crdPath, outputPath)
	require.NoError(t, err)

	// Verify output file exists
	_, err = os.Stat(outputPath)
	require.NoError(t, err, "Output file should exist")

	// Read and parse the JSON Schema
	schemaBytes, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	var schema map[string]interface{}
	err = json.Unmarshal(schemaBytes, &schema)
	require.NoError(t, err)

	// Verify JSON Schema metadata
	assert.Equal(t, "http://json-schema.org/draft-07/schema#", schema["$schema"])

	// Verify spec properties exist
	properties, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok, "properties should be a map")
	
	_, hasSpec := properties["spec"]
	assert.True(t, hasSpec, "spec property should exist")

	// Verify metadata was removed from properties
	_, hasMetadata := properties["metadata"]
	assert.False(t, hasMetadata, "metadata property should be removed")
}

func TestGenerateFromCRD_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	crdPath := filepath.Join(tmpDir, "nonexistent.yaml")
	outputPath := filepath.Join(tmpDir, "schema.json")

	err := GenerateFromCRD(crdPath, outputPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read CRD file")
}

func TestGenerateFromCRD_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	crdPath := filepath.Join(tmpDir, "invalid.yaml")
	outputPath := filepath.Join(tmpDir, "schema.json")

	// Write invalid YAML
	err := os.WriteFile(crdPath, []byte("invalid: yaml: content: [[["), 0644)
	require.NoError(t, err)

	err = GenerateFromCRD(crdPath, outputPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse CRD YAML")
}

func TestGenerateFromCRD_NoSchema(t *testing.T) {
	tmpDir := t.TempDir()
	crdPath := filepath.Join(tmpDir, "crd.yaml")
	outputPath := filepath.Join(tmpDir, "schema.json")

	// Create a CRD with no schema
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
					// No Schema field
				},
			},
		},
	}

	data, err := yaml.Marshal(crd)
	require.NoError(t, err)
	err = os.WriteFile(crdPath, data, 0644)
	require.NoError(t, err)

	err = GenerateFromCRD(crdPath, outputPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no schema found in CRD")
}

func TestGenerateFromCRD_WithKubernetesExtensions(t *testing.T) {
	tmpDir := t.TempDir()
	crdPath := filepath.Join(tmpDir, "crd.yaml")
	outputPath := filepath.Join(tmpDir, "schema.json")

	// Create a CRD with x-kubernetes-* extensions
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
								"spec": {
									Type: "object",
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										"field": {
											Type:        "string",
											Description: "test field",
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

	// Generate JSON Schema
	err = GenerateFromCRD(crdPath, outputPath)
	require.NoError(t, err)

	// Read and verify the schema doesn't have kubernetes extensions
	schemaBytes, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	schemaStr := string(schemaBytes)
	assert.NotContains(t, schemaStr, "x-kubernetes-")
}

func TestRemoveKubernetesExtensions_Map(t *testing.T) {
	obj := map[string]interface{}{
		"type": "object",
		"x-kubernetes-preserve-unknown-fields": true,
		"x-kubernetes-embedded-resource":       true,
		"properties": map[string]interface{}{
			"field1": map[string]interface{}{
				"type":                           "string",
				"x-kubernetes-list-map-keys":     []string{"name"},
				"x-kubernetes-some-other-field":  "value",
			},
		},
	}

	removeKubernetesExtensions(obj)

	// Check that x-kubernetes-* fields are removed at top level
	_, hasPreserve := obj["x-kubernetes-preserve-unknown-fields"]
	assert.False(t, hasPreserve, "x-kubernetes-preserve-unknown-fields should be removed")
	
	_, hasEmbedded := obj["x-kubernetes-embedded-resource"]
	assert.False(t, hasEmbedded, "x-kubernetes-embedded-resource should be removed")

	// Check that x-kubernetes-* fields are removed in nested properties
	props := obj["properties"].(map[string]interface{})
	field1 := props["field1"].(map[string]interface{})
	
	_, hasListMapKeys := field1["x-kubernetes-list-map-keys"]
	assert.False(t, hasListMapKeys, "x-kubernetes-list-map-keys should be removed")
	
	_, hasSomeOther := field1["x-kubernetes-some-other-field"]
	assert.False(t, hasSomeOther, "x-kubernetes-some-other-field should be removed")

	// Check that normal fields are preserved
	assert.Equal(t, "object", obj["type"])
	assert.Equal(t, "string", field1["type"])
}

func TestRemoveKubernetesExtensions_Array(t *testing.T) {
	obj := []interface{}{
		map[string]interface{}{
			"type":                      "string",
			"x-kubernetes-list-type":    "atomic",
		},
		map[string]interface{}{
			"type":                      "integer",
			"x-kubernetes-int-or-string": true,
		},
	}

	removeKubernetesExtensions(obj)

	// Check first item
	item0 := obj[0].(map[string]interface{})
	_, hasListType := item0["x-kubernetes-list-type"]
	assert.False(t, hasListType, "x-kubernetes-list-type should be removed")
	assert.Equal(t, "string", item0["type"])

	// Check second item
	item1 := obj[1].(map[string]interface{})
	_, hasIntOrString := item1["x-kubernetes-int-or-string"]
	assert.False(t, hasIntOrString, "x-kubernetes-int-or-string should be removed")
	assert.Equal(t, "integer", item1["type"])
}

func TestRemoveKubernetesExtensions_NonMapNonArray(_ *testing.T) {
	// Should not panic or error on primitive values
	removeKubernetesExtensions("string")
	removeKubernetesExtensions(42)
	removeKubernetesExtensions(true)
	removeKubernetesExtensions(nil)
	// Just testing that it doesn't crash
}

func TestConvertToJSONSchema_WithMultipleVersions(t *testing.T) {
	tmpDir := t.TempDir()
	crdPath := filepath.Join(tmpDir, "crd.yaml")
	outputPath := filepath.Join(tmpDir, "schema.json")

	// Create a CRD with multiple versions (should use first one with schema)
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
					// No schema
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
										"version": {Type: "string"},
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

	// Generate JSON Schema
	err = GenerateFromCRD(crdPath, outputPath)
	require.NoError(t, err)

	// Verify it used the v1 schema
	schemaBytes, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	var schema map[string]interface{}
	err = json.Unmarshal(schemaBytes, &schema)
	require.NoError(t, err)

	properties := schema["properties"].(map[string]interface{})
	spec := properties["spec"].(map[string]interface{})
	specProps := spec["properties"].(map[string]interface{})
	
	_, hasVersion := specProps["version"]
	assert.True(t, hasVersion, "version field from v1 schema should exist")
}

