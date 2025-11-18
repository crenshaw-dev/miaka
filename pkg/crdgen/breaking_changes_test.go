package crdgen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func TestCheckBreakingChanges_NoExistingCRD(t *testing.T) {
	// Test that when no existing CRD exists, no error is returned
	fs := afero.NewMemMapFs()

	newCRD := []byte(`apiVersion: apiextensions.k8s.io/v1
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
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              replicas:
                type: integer
`)

	err := CheckBreakingChangesWithFS(fs, "nonexistent.yaml", newCRD)
	if err != nil {
		t.Errorf("Expected no error when old CRD doesn't exist, got: %v", err)
	}
}

func TestCheckBreakingChanges_CompatibleChange(t *testing.T) {
	// Test that compatible changes (adding a field) don't cause errors
	fs := afero.NewMemMapFs()

	oldCRD := []byte(`apiVersion: apiextensions.k8s.io/v1
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
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              replicas:
                type: integer
`)

	newCRD := []byte(`apiVersion: apiextensions.k8s.io/v1
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
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              replicas:
                type: integer
              newField:
                type: string
`)

	// Write old CRD to filesystem
	if err := afero.WriteFile(fs, "old.yaml", oldCRD, 0644); err != nil {
		t.Fatalf("Failed to write old CRD: %v", err)
	}

	err := CheckBreakingChangesWithFS(fs, "old.yaml", newCRD)
	if err != nil {
		t.Errorf("Expected no error for compatible change, got: %v", err)
	}
}

func TestCheckBreakingChanges_FieldTypeChange(t *testing.T) {
	// Test that changing a field type is detected as a breaking change
	fs := afero.NewMemMapFs()

	oldCRD := []byte(`apiVersion: apiextensions.k8s.io/v1
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
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              replicas:
                type: integer
`)

	newCRD := []byte(`apiVersion: apiextensions.k8s.io/v1
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
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              replicas:
                type: string
`)

	// Write old CRD to filesystem
	if err := afero.WriteFile(fs, "old.yaml", oldCRD, 0644); err != nil {
		t.Fatalf("Failed to write old CRD: %v", err)
	}

	err := CheckBreakingChangesWithFS(fs, "old.yaml", newCRD)
	if err == nil {
		t.Error("Expected error for type change, got none")
	}
}

func TestCheckBreakingChanges_FieldRemoval(t *testing.T) {
	// Test that removing a field is detected as a breaking change
	fs := afero.NewMemMapFs()

	oldCRD := []byte(`apiVersion: apiextensions.k8s.io/v1
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
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              replicas:
                type: integer
              image:
                type: string
`)

	newCRD := []byte(`apiVersion: apiextensions.k8s.io/v1
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
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              replicas:
                type: integer
`)

	// Write old CRD to filesystem
	if err := afero.WriteFile(fs, "old.yaml", oldCRD, 0644); err != nil {
		t.Fatalf("Failed to write old CRD: %v", err)
	}

	err := CheckBreakingChangesWithFS(fs, "old.yaml", newCRD)
	if err == nil {
		t.Error("Expected error for field removal, got none")
	}
}

func TestCheckBreakingChanges_RealFilesystem(t *testing.T) {
	// Test with real filesystem operations
	tmpDir, err := os.MkdirTemp("", "crdgen-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldCRDPath := filepath.Join(tmpDir, "old.yaml")
	oldCRD := []byte(`apiVersion: apiextensions.k8s.io/v1
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
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              replicas:
                type: integer
`)

	if err := os.WriteFile(oldCRDPath, oldCRD, 0644); err != nil {
		t.Fatalf("Failed to write old CRD: %v", err)
	}

	newCRD := []byte(`apiVersion: apiextensions.k8s.io/v1
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
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              replicas:
                type: integer
              newField:
                type: string
`)

	// Should not error for compatible change
	err = CheckBreakingChanges(oldCRDPath, newCRD)
	if err != nil {
		t.Errorf("Expected no error for compatible change, got: %v", err)
	}
}

func TestCheckBreakingChangesWithFiles(t *testing.T) {
	// Test the convenience function that takes two file paths
	tmpDir, err := os.MkdirTemp("", "crdgen-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldCRDPath := filepath.Join(tmpDir, "old.yaml")
	newCRDPath := filepath.Join(tmpDir, "new.yaml")

	oldCRD := []byte(`apiVersion: apiextensions.k8s.io/v1
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
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              replicas:
                type: integer
`)

	newCRD := []byte(`apiVersion: apiextensions.k8s.io/v1
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
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              replicas:
                type: string
`)

	if err := os.WriteFile(oldCRDPath, oldCRD, 0644); err != nil {
		t.Fatalf("Failed to write old CRD: %v", err)
	}

	if err := os.WriteFile(newCRDPath, newCRD, 0644); err != nil {
		t.Fatalf("Failed to write new CRD: %v", err)
	}

	// Should error for type change
	err = CheckBreakingChangesWithFiles(oldCRDPath, newCRDPath)
	if err == nil {
		t.Error("Expected error for type change, got none")
	}
}
