package validation

import (
	"os"
	"path/filepath"
	"testing"
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
	if err := os.WriteFile(crdPath, []byte(crdContent), 0644); err != nil {
		t.Fatalf("Failed to write CRD: %v", err)
	}

	// Create a valid resource
	resourceContent := `apiVersion: example.com/v1alpha1
kind: Example
metadata:
  name: test-example
replicas: 3
appName: "myapp"
`
	resourcePath := filepath.Join(tmpDir, "resource.yaml")
	if err := os.WriteFile(resourcePath, []byte(resourceContent), 0644); err != nil {
		t.Fatalf("Failed to write resource: %v", err)
	}

	// Validate - should succeed
	err := ValidateAgainstCRD(crdPath, resourcePath)
	if err != nil {
		t.Errorf("Expected validation to succeed, got error: %v", err)
	}
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
	if err := os.WriteFile(crdPath, []byte(crdContent), 0644); err != nil {
		t.Fatalf("Failed to write CRD: %v", err)
	}

	// Create an invalid resource (replicas is a string instead of integer)
	resourceContent := `apiVersion: example.com/v1alpha1
kind: Example
metadata:
  name: test-example
replicas: "three"
appName: "myapp"
`
	resourcePath := filepath.Join(tmpDir, "resource.yaml")
	if err := os.WriteFile(resourcePath, []byte(resourceContent), 0644); err != nil {
		t.Fatalf("Failed to write resource: %v", err)
	}

	// Validate - should fail
	err := ValidateAgainstCRD(crdPath, resourcePath)
	if err == nil {
		t.Error("Expected validation to fail with type error, but it succeeded")
	}
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
	if err := os.WriteFile(crdPath, []byte(crdContent), 0644); err != nil {
		t.Fatalf("Failed to write CRD: %v", err)
	}

	// Create an invalid resource (replicas violates minimum constraint)
	resourceContent := `apiVersion: example.com/v1alpha1
kind: Example
metadata:
  name: test-example
replicas: 0
appName: "myapp"
`
	resourcePath := filepath.Join(tmpDir, "resource.yaml")
	if err := os.WriteFile(resourcePath, []byte(resourceContent), 0644); err != nil {
		t.Fatalf("Failed to write resource: %v", err)
	}

	// Validate - should fail
	err := ValidateAgainstCRD(crdPath, resourcePath)
	if err == nil {
		t.Error("Expected validation to fail with minimum constraint violation, but it succeeded")
	}
}

// TestValidateAgainstCRD_MalformedCRD tests error handling for malformed CRD
func TestValidateAgainstCRD_MalformedCRD(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a malformed CRD
	crdContent := `this is not valid yaml: [[[`
	crdPath := filepath.Join(tmpDir, "crd.yaml")
	if err := os.WriteFile(crdPath, []byte(crdContent), 0644); err != nil {
		t.Fatalf("Failed to write CRD: %v", err)
	}

	// Create a valid resource
	resourceContent := `apiVersion: example.com/v1alpha1
kind: Example
metadata:
  name: test-example
replicas: 3
`
	resourcePath := filepath.Join(tmpDir, "resource.yaml")
	if err := os.WriteFile(resourcePath, []byte(resourceContent), 0644); err != nil {
		t.Fatalf("Failed to write resource: %v", err)
	}

	// Validate - should fail with unmarshal error
	err := ValidateAgainstCRD(crdPath, resourcePath)
	if err == nil {
		t.Error("Expected validation to fail with malformed CRD, but it succeeded")
	}
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
	if err := os.WriteFile(crdPath, []byte(crdContent), 0644); err != nil {
		t.Fatalf("Failed to write CRD: %v", err)
	}

	// Create a malformed resource
	resourceContent := `this is not valid yaml: [[[`
	resourcePath := filepath.Join(tmpDir, "resource.yaml")
	if err := os.WriteFile(resourcePath, []byte(resourceContent), 0644); err != nil {
		t.Fatalf("Failed to write resource: %v", err)
	}

	// Validate - should fail with unmarshal error
	err := ValidateAgainstCRD(crdPath, resourcePath)
	if err == nil {
		t.Error("Expected validation to fail with malformed resource, but it succeeded")
	}
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
	if err := os.WriteFile(resourcePath, []byte(resourceContent), 0644); err != nil {
		t.Fatalf("Failed to write resource: %v", err)
	}

	// Validate with non-existent CRD file
	crdPath := filepath.Join(tmpDir, "nonexistent.yaml")
	err := ValidateAgainstCRD(crdPath, resourcePath)
	if err == nil {
		t.Error("Expected validation to fail with missing CRD file, but it succeeded")
	}
}
