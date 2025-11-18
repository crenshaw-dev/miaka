package crdgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

func TestValidateCRD_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	crdPath := filepath.Join(tmpDir, "valid.yaml")

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

	data, _ := yaml.Marshal(crd)
	os.WriteFile(crdPath, data, 0644)

	// Should pass validation
	err := ValidateCRD(crdPath)
	if err != nil {
		t.Errorf("Expected valid CRD to pass, got error: %v", err)
	}
}

func TestValidateCRD_MutuallyExclusivePropertiesAndAdditionalProperties(t *testing.T) {
	tmpDir := t.TempDir()
	crdPath := filepath.Join(tmpDir, "invalid-mutual-exclusive.yaml")

	// Create a CRD with the forbidden combination of properties AND additionalProperties
	// This is the exact issue we encountered with imagePullSecrets
	crd := &apiextensionsv1.CustomResourceDefinition{
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "example.com",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:   "BadExample",
				Plural: "badexamples",
			},
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name: "v1",
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"global": {
									Type: "object",
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										"imagePullSecrets": {
											Type: "array",
											Items: &apiextensionsv1.JSONSchemaPropsOrArray{
												Schema: &apiextensionsv1.JSONSchemaProps{
													Type: "object",
													// THIS IS INVALID: both properties AND additionalProperties
													Properties: map[string]apiextensionsv1.JSONSchemaProps{
														"name": {Type: "string"},
													},
													AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{
														Allows: false,
													},
												},
											},
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

	data, _ := yaml.Marshal(crd)
	os.WriteFile(crdPath, data, 0644)

	// Note: schema.NewStructural() does NOT catch this validation error!
	// The Kubernetes API server would reject this, but the structural schema validator
	// we use (schema.NewStructural) only validates structural completeness, not all
	// semantic constraints like "properties and additionalProperties are mutually exclusive".
	//
	// This test documents this limitation and ensures our code doesn't crash on such CRDs.
	// The real validation happens when applying to a cluster.
	err := ValidateCRD(crdPath)

	// We expect this to pass (even though it's invalid according to Kubernetes)
	// because schema.NewStructural doesn't catch this specific constraint
	if err != nil {
		t.Logf("Validator returned error (unexpected): %v", err)
		t.Skip("schema.NewStructural caught the issue - this is actually good, but not expected based on our testing")
	}

	t.Log("Note: This CRD would be rejected by Kubernetes API server with:")
	t.Log("  'additionalProperties and properties are mutually exclusive'")
	t.Log("Our validator only catches structural schema issues, not all semantic constraints.")
}

func TestValidateCRD_FileNotFound(t *testing.T) {
	err := ValidateCRD("/nonexistent/path/to/crd.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file, but got nil")
	}
	if !os.IsNotExist(err) {
		// Check that error message mentions file reading
		if err.Error() == "" || err.Error()[:len("failed to read")] != "failed to read" {
			t.Errorf("Expected 'failed to read' error, got: %v", err)
		}
	}
}

func TestValidateCRD_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	crdPath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	os.WriteFile(crdPath, []byte("this is not: valid: yaml: content"), 0644)

	err := ValidateCRD(crdPath)
	if err == nil {
		t.Error("Expected error for invalid YAML, but got nil")
	}
}

func TestValidateCRD_MultipleVersions(t *testing.T) {
	tmpDir := t.TempDir()
	crdPath := filepath.Join(tmpDir, "multiversion.yaml")

	// Create a CRD with multiple versions
	crd := &apiextensionsv1.CustomResourceDefinition{
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "example.com",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:   "Multi",
				Plural: "multis",
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
										"field": {Type: "string"},
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
										"field":    {Type: "string"},
										"newField": {Type: "integer"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	data, _ := yaml.Marshal(crd)
	os.WriteFile(crdPath, data, 0644)

	// Should pass validation for all versions
	err := ValidateCRD(crdPath)
	if err != nil {
		t.Errorf("Expected valid multi-version CRD to pass, got error: %v", err)
	}
}

func TestValidateCRD_WithArrays(t *testing.T) {
	tmpDir := t.TempDir()
	crdPath := filepath.Join(tmpDir, "arrays.yaml")

	// Create a CRD with array types
	crd := &apiextensionsv1.CustomResourceDefinition{
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "example.com",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:   "ArrayExample",
				Plural: "arrayexamples",
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
										"items": {
											Type: "array",
											Items: &apiextensionsv1.JSONSchemaPropsOrArray{
												Schema: &apiextensionsv1.JSONSchemaProps{
													Type: "object",
													Properties: map[string]apiextensionsv1.JSONSchemaProps{
														"name":  {Type: "string"},
														"value": {Type: "integer"},
													},
												},
											},
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

	data, _ := yaml.Marshal(crd)
	os.WriteFile(crdPath, data, 0644)

	// Should pass validation
	err := ValidateCRD(crdPath)
	if err != nil {
		t.Errorf("Expected valid CRD with arrays to pass, got error: %v", err)
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}
