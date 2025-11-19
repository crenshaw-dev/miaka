package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateAgainstCRD_Valid tests validation with a valid resource
func TestValidateAgainstCRD_Valid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a sample CRD
	crdContent := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: examples.example.com
spec:
  group: example.com
  names:
    kind: Example
    plural: examples
  scope: Namespaced
  versions:
  - name: v1alpha1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          replicas:
            type: integer
            minimum: 1
          appName:
            type: string
`
	crdPath := filepath.Join(tmpDir, "crd.yaml")
	err := os.WriteFile(crdPath, []byte(crdContent), 0644)
	require.NoError(t, err, "Failed to write CRD")

	// Create a valid resource
	resourceContent := `apiVersion: example.com/v1alpha1
kind: Example
metadata:
  name: test-example
replicas: 3
appName: "myapp"
`
	resourcePath := filepath.Join(tmpDir, "resource.yaml")
	err = os.WriteFile(resourcePath, []byte(resourceContent), 0644)
	require.NoError(t, err, "Failed to write resource")

	// Validate - should succeed
	err = ValidateAgainstCRD(crdPath, resourcePath)
	assert.NoError(t, err, "Expected validation to succeed")
}

// TestValidateAgainstCRD_InvalidType tests validation with wrong type
func TestValidateAgainstCRD_InvalidType(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a sample CRD
	crdContent := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: examples.example.com
spec:
  group: example.com
  names:
    kind: Example
    plural: examples
  scope: Namespaced
  versions:
  - name: v1alpha1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          replicas:
            type: integer
            minimum: 1
          appName:
            type: string
`
	crdPath := filepath.Join(tmpDir, "crd.yaml")
	err := os.WriteFile(crdPath, []byte(crdContent), 0644)
	require.NoError(t, err, "Failed to write CRD")

	// Create an invalid resource (replicas is a string instead of integer)
	resourceContent := `apiVersion: example.com/v1alpha1
kind: Example
metadata:
  name: test-example
replicas: "three"
appName: "myapp"
`
	resourcePath := filepath.Join(tmpDir, "resource.yaml")
	err = os.WriteFile(resourcePath, []byte(resourceContent), 0644)
	require.NoError(t, err, "Failed to write resource")

	// Validate - should fail
	err = ValidateAgainstCRD(crdPath, resourcePath)
	assert.Error(t, err, "Expected validation to fail with type error")
}

// TestValidateAgainstCRD_ValidationConstraint tests validation with constraint violation
func TestValidateAgainstCRD_ValidationConstraint(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a sample CRD with minimum constraint
	crdContent := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: examples.example.com
spec:
  group: example.com
  names:
    kind: Example
    plural: examples
  scope: Namespaced
  versions:
  - name: v1alpha1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          replicas:
            type: integer
            minimum: 1
          appName:
            type: string
`
	crdPath := filepath.Join(tmpDir, "crd.yaml")
	err := os.WriteFile(crdPath, []byte(crdContent), 0644)
	require.NoError(t, err, "Failed to write CRD")

	// Create an invalid resource (replicas violates minimum constraint)
	resourceContent := `apiVersion: example.com/v1alpha1
kind: Example
metadata:
  name: test-example
replicas: 0
appName: "myapp"
`
	resourcePath := filepath.Join(tmpDir, "resource.yaml")
	err = os.WriteFile(resourcePath, []byte(resourceContent), 0644)
	require.NoError(t, err, "Failed to write resource")

	// Validate - should fail
	err = ValidateAgainstCRD(crdPath, resourcePath)
	assert.Error(t, err, "Expected validation to fail with minimum constraint violation")
}

// TestValidateAgainstCRD_MalformedCRD tests error handling for malformed CRD
func TestValidateAgainstCRD_MalformedCRD(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a malformed CRD
	crdContent := `this is not valid yaml: [[[`
	crdPath := filepath.Join(tmpDir, "crd.yaml")
	err := os.WriteFile(crdPath, []byte(crdContent), 0644)
	require.NoError(t, err, "Failed to write CRD")

	// Create a valid resource
	resourceContent := `apiVersion: example.com/v1alpha1
kind: Example
metadata:
  name: test-example
replicas: 3
`
	resourcePath := filepath.Join(tmpDir, "resource.yaml")
	err = os.WriteFile(resourcePath, []byte(resourceContent), 0644)
	require.NoError(t, err, "Failed to write resource")

	// Validate - should fail with unmarshal error
	err = ValidateAgainstCRD(crdPath, resourcePath)
	assert.Error(t, err, "Expected validation to fail with malformed CRD")
}

// TestValidateAgainstCRD_MalformedResource tests error handling for malformed resource
func TestValidateAgainstCRD_MalformedResource(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid CRD
	crdContent := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: examples.example.com
spec:
  group: example.com
  names:
    kind: Example
    plural: examples
  scope: Namespaced
  versions:
  - name: v1alpha1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          replicas:
            type: integer
`
	crdPath := filepath.Join(tmpDir, "crd.yaml")
	err := os.WriteFile(crdPath, []byte(crdContent), 0644)
	require.NoError(t, err, "Failed to write CRD")

	// Create a malformed resource
	resourceContent := `this is not valid yaml: [[[`
	resourcePath := filepath.Join(tmpDir, "resource.yaml")
	err = os.WriteFile(resourcePath, []byte(resourceContent), 0644)
	require.NoError(t, err, "Failed to write resource")

	// Validate - should fail with unmarshal error
	err = ValidateAgainstCRD(crdPath, resourcePath)
	assert.Error(t, err, "Expected validation to fail with malformed resource")
}

// TestValidateAgainstCRD_MissingCRDFile tests error handling for missing CRD file
func TestValidateAgainstCRD_MissingCRDFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid resource
	resourceContent := `apiVersion: example.com/v1alpha1
kind: Example
metadata:
  name: test-example
replicas: 3
`
	resourcePath := filepath.Join(tmpDir, "resource.yaml")
	err := os.WriteFile(resourcePath, []byte(resourceContent), 0644)
	require.NoError(t, err, "Failed to write resource")

	// Validate with non-existent CRD file
	crdPath := filepath.Join(tmpDir, "nonexistent.yaml")
	err = ValidateAgainstCRD(crdPath, resourcePath)
	assert.Error(t, err, "Expected validation to fail with missing CRD file")
}
