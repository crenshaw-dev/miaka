package gotypes

import (
	"strings"
	"testing"

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
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	output := string(code)

	// Check package declaration
	if !strings.Contains(output, "package v1alpha1") {
		t.Error("Expected package v1alpha1")
	}

	// Check imports
	if !strings.Contains(output, "metav1") {
		t.Error("Expected metav1 import")
	}

	// Check main type
	if !strings.Contains(output, "type Simple struct") {
		t.Error("Expected Simple type")
	}

	// Check Spec struct
	if !strings.Contains(output, "type SimpleSpec struct") {
		t.Error("Expected SimpleSpec type")
	}

	// Check fields
	if !strings.Contains(output, "Name string") {
		t.Error("Expected Name field")
	}
	if !strings.Contains(output, "Count int") {
		t.Error("Expected Count field")
	}

	// Check comments
	if !strings.Contains(output, "// The name field") {
		t.Error("Expected name field comment")
	}
	if !strings.Contains(output, "// The count field") {
		t.Error("Expected count field comment")
	}
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
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	output := string(code)

	// Check slice field
	if !strings.Contains(output, "Items []ItemConfig") {
		t.Error("Expected Items []ItemConfig field")
	}

	// Check nested struct
	if !strings.Contains(output, "type ItemConfig struct") {
		t.Error("Expected ItemConfig type")
	}
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
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	output := string(code)

	// Check kubebuilder tags are preserved
	if !strings.Contains(output, "+kubebuilder:validation:Minimum=1") {
		t.Error("Expected kubebuilder Minimum tag")
	}
	if !strings.Contains(output, "+kubebuilder:validation:Maximum=65535") {
		t.Error("Expected kubebuilder Maximum tag")
	}
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
			if result != tt.expected {
				t.Errorf("generateStructDescription(%q) = %q, want %q", tt.input, result, tt.expected)
			}
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
			if result != tt.expected {
				t.Errorf("fixInlineComments() mismatch\nGot:\n%s\nWant:\n%s", result, tt.expected)
			}
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
			if result != tt.expected {
				t.Errorf("getIndent(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

