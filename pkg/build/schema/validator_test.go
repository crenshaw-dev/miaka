package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSchema_NoInterfaceTypes(t *testing.T) {
	s := &Schema{
		APIVersion: "example.com/v1",
		Kind:       "Example",
		Package:    "v1",
		Structs: []StructDef{
			{
				Name: "ExampleSpec",
				Fields: []Field{
					{
						Name:     "Name",
						JSONName: "name",
						Type:     "string",
						YAMLPath: "name",
						Line:     1,
					},
					{
						Name:     "Count",
						JSONName: "count",
						Type:     "int",
						YAMLPath: "count",
						Line:     2,
					},
					{
						Name:     "Enabled",
						JSONName: "enabled",
						Type:     "bool",
						YAMLPath: "enabled",
						Line:     3,
					},
				},
			},
		},
	}

	err := ValidateSchema(s)
	assert.NoError(t, err, "ValidateSchema() expected no error for valid schema")
}

func TestValidateSchema_InterfaceType(t *testing.T) {
	s := &Schema{
		APIVersion: "example.com/v1",
		Kind:       "Example",
		Package:    "v1",
		Structs: []StructDef{
			{
				Name: "ExampleSpec",
				Fields: []Field{
					{
						Name:     "Name",
						JSONName: "name",
						Type:     "string",
						YAMLPath: "name",
						Line:     1,
					},
					{
						Name:     "Data",
						JSONName: "data",
						Type:     "interface{}",
						YAMLPath: "data",
						Line:     2,
					},
				},
			},
		},
	}

	err := ValidateSchema(s)
	require.Error(t, err, "ValidateSchema() expected error for interface{} type")

	// Check that error is of correct type
	interfaceErr, ok := err.(*InterfaceTypeError)
	require.True(t, ok, "ValidateSchema() expected *InterfaceTypeError, got %T", err)

	assert.Len(t, interfaceErr.Fields, 1, "Expected 1 interface field")

	// Check error message contains helpful information
	errMsg := err.Error()
	assert.Contains(t, errMsg, "interface{}", "Error message should mention interface{}")
	assert.Contains(t, errMsg, "data", "Error message should mention the field name")
}

func TestValidateSchema_InterfaceSlice(t *testing.T) {
	s := &Schema{
		APIVersion: "example.com/v1",
		Kind:       "Example",
		Package:    "v1",
		Structs: []StructDef{
			{
				Name: "ExampleSpec",
				Fields: []Field{
					{
						Name:     "Items",
						JSONName: "items",
						Type:     "[]interface{}",
						YAMLPath: "items",
						Line:     1,
						IsSlice:  false,
					},
				},
			},
		},
	}

	err := ValidateSchema(s)
	require.Error(t, err, "ValidateSchema() expected error for []interface{} type")

	interfaceErr, ok := err.(*InterfaceTypeError)
	require.True(t, ok, "ValidateSchema() expected *InterfaceTypeError, got %T", err)

	assert.Len(t, interfaceErr.Fields, 1, "Expected 1 interface field")

	assert.True(t, interfaceErr.Fields[0].IsArray, "Expected IsArray to be true for []interface{}")
}

func TestValidateSchema_InterfaceSliceElement(t *testing.T) {
	s := &Schema{
		APIVersion: "example.com/v1",
		Kind:       "Example",
		Package:    "v1",
		Structs: []StructDef{
			{
				Name: "ExampleSpec",
				Fields: []Field{
					{
						Name:     "Items",
						JSONName: "items",
						Type:     "[]MyItem",
						ElemType: "interface{}",
						IsSlice:  true,
						YAMLPath: "items",
						Line:     1,
					},
				},
			},
		},
	}

	err := ValidateSchema(s)
	require.Error(t, err, "ValidateSchema() expected error for slice with interface{} element")

	interfaceErr, ok := err.(*InterfaceTypeError)
	require.True(t, ok, "ValidateSchema() expected *InterfaceTypeError, got %T", err)

	assert.Len(t, interfaceErr.Fields, 1, "Expected 1 interface field")

	assert.True(t, interfaceErr.Fields[0].IsArray, "Expected IsArray to be true for slice with interface{} element")
}

func TestValidateSchema_MultipleInterfaceTypes(t *testing.T) {
	s := &Schema{
		APIVersion: "example.com/v1",
		Kind:       "Example",
		Package:    "v1",
		Structs: []StructDef{
			{
				Name: "ExampleSpec",
				Fields: []Field{
					{
						Name:     "Data",
						JSONName: "data",
						Type:     "interface{}",
						YAMLPath: "data",
						Line:     1,
					},
					{
						Name:     "Config",
						JSONName: "config",
						Type:     "interface{}",
						YAMLPath: "config",
						Line:     2,
					},
					{
						Name:     "Items",
						JSONName: "items",
						Type:     "[]interface{}",
						YAMLPath: "items",
						Line:     3,
					},
				},
			},
		},
	}

	err := ValidateSchema(s)
	require.Error(t, err, "ValidateSchema() expected error for multiple interface{} types")

	interfaceErr, ok := err.(*InterfaceTypeError)
	require.True(t, ok, "ValidateSchema() expected *InterfaceTypeError, got %T", err)

	assert.Len(t, interfaceErr.Fields, 3, "Expected 3 interface fields")
}

func TestValidateSchema_EmptySchema(t *testing.T) {
	s := &Schema{
		APIVersion: "example.com/v1",
		Kind:       "Example",
		Package:    "v1",
		Structs:    []StructDef{},
	}

	err := ValidateSchema(s)
	assert.NoError(t, err, "ValidateSchema() expected no error for empty schema")
}

func TestInterfaceTypeError_ErrorMessage(t *testing.T) {
	err := &InterfaceTypeError{
		Fields: []InterfaceTypeLocation{
			{
				StructName: "ExampleSpec",
				FieldName:  "Data",
				FieldPath:  "ExampleSpec.data",
				YAMLPath:   "data",
				Line:       5,
				IsArray:    false,
			},
			{
				StructName: "ExampleSpec",
				FieldName:  "Items",
				FieldPath:  "ExampleSpec.items",
				YAMLPath:   "items",
				Line:       10,
				IsArray:    true,
			},
		},
	}

	errMsg := err.Error()

	// Check that error message contains key information
	requiredStrings := []string{
		"interface{}",
		"2 field(s)",
		"Line 5: data",
		"Line 10: items",
		"empty array",
		"Solutions",
	}

	for _, required := range requiredStrings {
		assert.Contains(t, errMsg, required, "Error message missing required string %q", required)
	}
}
