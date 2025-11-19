package gotypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/crenshaw-dev/miaka/pkg/build/schema"
)

func TestGenerate_SimpleStruct(t *testing.T) {
	schema := &schema.Schema{
		APIVersion: "example.com/v1alpha1",
		Kind:       "Simple",
		Package:    "v1alpha1",
		Structs: []schema.StructDef{
			{
				Name: "SimpleSpec",
				Comments: []string{
					"Simple specification",
				},
				Fields: []schema.Field{
					{
						Name:     "Name",
						JSONName: "name",
						Type:     "string",
						Comments: []string{"The name field"},
					},
					{
						Name:     "Count",
						JSONName: "count",
						Type:     "int",
						Comments: []string{"The count field"},
					},
				},
			},
		},
	}

	g := NewGenerator(schema)
	code, err := g.Generate()
	require.NoError(t, err, "Generate() failed: %v")

	output := string(code)

	// Check package declaration
	assert.Contains(t, output, "package v1alpha1", "Expected package v1alpha1")

	// Check imports
	assert.Contains(t, output, "metav1", "Expected metav1 import")

	// Check main type
	assert.Contains(t, output, "type Simple struct", "Expected Simple type")

	// Check Spec struct
	assert.Contains(t, output, "type SimpleSpec struct", "Expected SimpleSpec type")

	// Check fields
	assert.Contains(t, output, "Name string", "Expected Name field")
	assert.Contains(t, output, "Count int", "Expected Count field")

	// Check comments
	assert.Contains(t, output, "// The name field", "Expected name field comment")
	assert.Contains(t, output, "// The count field", "Expected count field comment")
}

func TestGenerate_WithSlice(t *testing.T) {
	schema := &schema.Schema{
		APIVersion: "example.com/v1alpha1",
		Kind:       "ListExample",
		Package:    "v1alpha1",
		Structs: []schema.StructDef{
			{
				Name:     "ItemConfig",
				Comments: []string{"item configuration"},
				Fields: []schema.Field{
					{
						Name:     "Value",
						JSONName: "value",
						Type:     "string",
						Comments: []string{"Item value"},
					},
				},
			},
			{
				Name: "ListExampleSpec",
				Fields: []schema.Field{
					{
						Name:     "Items",
						JSONName: "items",
						IsSlice:  true,
						ElemType: "ItemConfig",
						Comments: []string{"List of items"},
					},
				},
			},
		},
	}

	g := NewGenerator(schema)
	code, err := g.Generate()
	require.NoError(t, err, "Generate() failed: %v")

	output := string(code)

	// Check slice field
	assert.Contains(t, output, "Items []ItemConfig", "Expected Items []ItemConfig field")

	// Check nested struct
	assert.Contains(t, output, "type ItemConfig struct", "Expected ItemConfig type")
}

func TestGenerate_WithKubebuilderTags(t *testing.T) {
	schema := &schema.Schema{
		APIVersion: "example.com/v1alpha1",
		Kind:       "Validated",
		Package:    "v1alpha1",
		Structs: []schema.StructDef{
			{
				Name: "ValidatedSpec",
				Fields: []schema.Field{
					{
						Name:     "Port",
						JSONName: "port",
						Type:     "int",
						Comments: []string{
							"Port number",
							"+kubebuilder:validation:Minimum=1",
							"+kubebuilder:validation:Maximum=65535",
						},
					},
				},
			},
		},
	}

	g := NewGenerator(schema)
	code, err := g.Generate()
	require.NoError(t, err, "Generate() failed: %v")

	output := string(code)

	// Check kubebuilder tags are preserved
	assert.Contains(t, output, "+kubebuilder:validation:Minimum=1", "Expected kubebuilder Minimum tag")
	assert.Contains(t, output, "+kubebuilder:validation:Maximum=65535", "Expected kubebuilder Maximum tag")
}

func TestGenerateStructDescription(t *testing.T) {
	g := &Generator{
		schema: &schema.Schema{
			Kind: "Example",
		},
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Config suffix",
			input:    "ServiceConfig",
			expected: "service configuration",
		},
		{
			name:     "Simple name",
			input:    "User",
			expected: "user",
		},
		{
			name:     "Multiple words",
			input:    "DatabaseConnection",
			expected: "database connection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.generateStructDescription(tt.input)
			assert.Equal(t, tt.expected, result, "generateStructDescription(%q) = %q, want %q")
		})
	}
}

func TestSplitPascalCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple",
			input:    "Service",
			expected: []string{"Service"},
		},
		{
			name:     "Two words",
			input:    "ServiceConfig",
			expected: []string{"Service", "Config"},
		},
		{
			name:     "Three words",
			input:    "DatabaseConnectionPool",
			expected: []string{"Database", "Connection", "Pool"},
		},
		{
			name:     "With acronym",
			input:    "HTTPServer",
			expected: []string{"HTTP", "Server"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitPascalCase(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitPascalCase(%q) returned %d words, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("splitPascalCase(%q)[%d] = %q, want %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestFixInlineComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Struct with inline comment",
			input:    "type Foo struct {// A comment\n\tBar int `json:\"bar,omitempty\"`\n}",
			expected: "type Foo struct {\n\t// A comment\n\tBar int `json:\"bar,omitempty\"`\n}",
		},
		{
			name:     "Field with inline comment",
			input:    "\tBar int `json:\"bar,omitempty\"`// Next field comment\n\tBaz string `json:\"baz,omitempty\"`",
			expected: "\tBar int `json:\"bar,omitempty\"`\n\n\t// Next field comment\n\tBaz string `json:\"baz,omitempty\"`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fixInlineComments(tt.input)
			assert.Equal(t, tt.expected, result, "fixInlineComments() mismatch\nGot:\n%s\nWant:\n%s")
		})
	}
}

func TestGetIndent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No indent",
			input:    "package main",
			expected: "",
		},
		{
			name:     "One tab",
			input:    "\tfield int",
			expected: "\t",
		},
		{
			name:     "Two tabs",
			input:    "\t\tfield int",
			expected: "\t\t",
		},
		{
			name:     "Spaces",
			input:    "    field int",
			expected: "    ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIndent(tt.input)
			assert.Equal(t, tt.expected, result, "getIndent(%q) = %q, want %q")
		})
	}
}

