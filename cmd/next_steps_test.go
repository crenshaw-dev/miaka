package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitNextSteps(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create a simple values.yaml
	valuesContent := `replicas: 3
image: nginx:latest`
	err := os.WriteFile("values.yaml", []byte(valuesContent), 0644)
	require.NoError(t, err)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set flags for non-interactive mode
	initApiVersion = "myapp.io/v1"
	initKind = "MyApp"
	initOutput = "example.values.yaml"

	// Run init command
	err = runInit(nil, []string{"values.yaml"})

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify command succeeded
	require.NoError(t, err)

	// Verify next steps are shown
	assert.Contains(t, output, "Next steps:")
	assert.Contains(t, output, "miaka build")
	assert.Contains(t, output, "example values for all fields")
	assert.Contains(t, output, "kubebuilder markers")

	// Verify file was created
	_, err = os.Stat("example.values.yaml")
	assert.NoError(t, err)
}

func TestBuildFirstTimeNextSteps(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create a valid example.values.yaml
	exampleContent := `apiVersion: myapp.io/v1
kind: MyApp
replicas: 3
image: nginx:latest`
	err := os.WriteFile("example.values.yaml", []byte(exampleContent), 0644)
	require.NoError(t, err)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set flags
	buildCRDPath = "crd.yaml"
	buildSchemaPath = "values.schema.json"
	buildTypesPath = ""

	// Run build command
	err = runBuild(nil, []string{"example.values.yaml"})

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify command succeeded
	require.NoError(t, err)

	// Verify first-time message and next steps are shown
	assert.Contains(t, output, "Generated schemas for the first time!")
	assert.Contains(t, output, "Next steps:")
	assert.Contains(t, output, "miaka validate")
	assert.Contains(t, output, "Improve your schema")
	assert.Contains(t, output, "kubebuilder validation markers")
	assert.Contains(t, output, "field descriptions as comments")
	assert.Contains(t, output, "Commit the generated files to git")
	assert.Contains(t, output, "git add")
	assert.Contains(t, output, "breaking change detection")

	// Verify files were created
	_, err = os.Stat("crd.yaml")
	require.NoError(t, err)
	_, err = os.Stat("values.schema.json")
	require.NoError(t, err)
}

func TestBuildSubsequentRunNoNextSteps(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create a valid example.values.yaml
	exampleContent := `apiVersion: myapp.io/v1
kind: MyApp
replicas: 3
image: nginx:latest`
	err := os.WriteFile("example.values.yaml", []byte(exampleContent), 0644)
	require.NoError(t, err)

	// Create an existing CRD to simulate subsequent run
	existingCRD := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: myapps.myapp.io
spec:
  group: myapp.io
  names:
    kind: MyApp
    plural: myapps
  scope: Namespaced
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        description: "MyApp is the Schema for the myapps API"
        properties:
          apiVersion:
            type: string
            description: "APIVersion defines the versioned schema of this representation of an object.\nServers should convert recognized schemas to the latest internal value, and\nmay reject unrecognized values.\nMore info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources"
          kind:
            type: string
            description: "Kind is a string value representing the REST resource this object represents.\nServers may infer this from the endpoint the client submits requests to.\nCannot be updated.\nIn CamelCase.\nMore info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds"
          replicas:
            type: integer
          image:
            type: string`
	err = os.WriteFile("crd.yaml", []byte(existingCRD), 0644)
	require.NoError(t, err)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set flags
	buildCRDPath = "crd.yaml"
	buildSchemaPath = "values.schema.json"
	buildTypesPath = ""

	// Run build command
	err = runBuild(nil, []string{"example.values.yaml"})

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify command succeeded
	require.NoError(t, err)

	// Verify first-time message is NOT shown (this is a subsequent run)
	assert.NotContains(t, output, "Generated schemas for the first time!")
	assert.NotContains(t, strings.ToLower(output), "next steps:")
}
