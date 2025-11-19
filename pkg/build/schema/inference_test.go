package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInferType(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "int",
			value:    42,
			expected: "int",
		},
		{
			name:     "int64",
			value:    int64(42),
			expected: "int",
		},
		{
			name:     "float64 that is actually int",
			value:    float64(42),
			expected: "int",
		},
		{
			name:     "float64",
			value:    3.14,
			expected: "float64",
		},
		{
			name:     "string",
			value:    "hello",
			expected: "string",
		},
		{
			name:     "bool true",
			value:    true,
			expected: "bool",
		},
		{
			name:     "bool false",
			value:    false,
			expected: "bool",
		},
		{
			name:     "slice",
			value:    []string{"a", "b"},
			expected: "interface{}",
		},
		{
			name:     "map",
			value:    map[string]interface{}{"key": "value"},
			expected: "interface{}",
		},
		{
			name:     "nil",
			value:    nil,
			expected: "interface{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InferType(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "snake_case",
			input:    "hello_world",
			expected: "HelloWorld",
		},
		{
			name:     "kebab-case",
			input:    "hello-world",
			expected: "HelloWorld",
		},
		{
			name:     "dot.separated",
			input:    "hello.world",
			expected: "HelloWorld",
		},
		{
			name:     "slash/separated",
			input:    "hello/world",
			expected: "HelloWorld",
		},
		{
			name:     "colon:separated",
			input:    "hello:world",
			expected: "HelloWorld",
		},
		{
			name:     "mixed separators",
			input:    "hello_world.test-case",
			expected: "HelloWorldTestCase",
		},
		{
			name:     "already PascalCase",
			input:    "HelloWorld",
			expected: "HelloWorld",
		},
		{
			name:     "single word lowercase",
			input:    "hello",
			expected: "Hello",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "multiple separators in a row",
			input:    "hello__world",
			expected: "HelloWorld",
		},
		{
			name:     "apiVersion style",
			input:    "example.com/v1alpha1",
			expected: "ExampleComV1alpha1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPascalCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateStructName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "service",
			expected: "ServiceConfig",
		},
		{
			name:     "already has Config suffix",
			input:    "serviceConfig",
			expected: "ServiceConfig",
		},
		{
			name:     "already has Spec suffix",
			input:    "serviceSpec",
			expected: "ServiceSpec",
		},
		{
			name:     "snake_case",
			input:    "my_service",
			expected: "MyServiceConfig",
		},
		{
			name:     "kebab-case",
			input:    "my-service",
			expected: "MyServiceConfig",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateStructName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatComments(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "comments with # prefix",
			input:    []string{"# Hello", "# World"},
			expected: []string{"Hello", "World"},
		},
		{
			name:     "comments without # prefix",
			input:    []string{"Hello", "World"},
			expected: []string{"Hello", "World"},
		},
		{
			name:     "comments with whitespace",
			input:    []string{"  # Hello  ", "  # World  "},
			expected: []string{"Hello", "World"},
		},
		{
			name:  "empty comments",
			input: []string{"", "  ", "#"},
		},
		{
			name:     "mixed",
			input:    []string{"# Comment 1", "", "Comment 2", "  "},
			expected: []string{"Comment 1", "Comment 2"},
		},
		{
			name:  "nil input",
			input: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatComments(tt.input)
			if len(tt.expected) == 0 {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseAPIVersion(t *testing.T) {
	tests := []struct {
		name        string
		apiVersion  string
		expected    string
		expectError bool
	}{
		{
			name:        "valid with group",
			apiVersion:  "example.com/v1alpha1",
			expected:    "v1alpha1",
			expectError: false,
		},
		{
			name:        "valid with subgroup",
			apiVersion:  "apps.example.com/v1",
			expected:    "v1",
			expectError: false,
		},
		{
			name:        "core API (no group)",
			apiVersion:  "v1",
			expected:    "v1",
			expectError: false,
		},
		{
			name:        "beta version",
			apiVersion:  "example.com/v1beta1",
			expected:    "v1beta1",
			expectError: false,
		},
		{
			name:        "invalid format (multiple slashes)",
			apiVersion:  "example.com/sub/v1",
			expected:    "",
			expectError: true,
		},
		{
			name:        "empty string",
			apiVersion:  "",
			expected:    "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAPIVersion(tt.apiVersion)

			if tt.expectError {
				assert.Error(t, err, "ParseAPIVersion(%q) expected error", tt.apiVersion)
				return
			}

			assert.NoError(t, err, "ParseAPIVersion(%q) unexpected error", tt.apiVersion)
			assert.Equal(t, tt.expected, result)
		})
	}
}
