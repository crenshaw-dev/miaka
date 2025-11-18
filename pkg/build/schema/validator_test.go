package schema

import (
	"strings"
	"testing"
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
	if err != nil {
		t.Errorf("ValidateSchema() expected no error for valid schema, got: %v", err)
	}
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
	if err == nil {
		t.Error("ValidateSchema() expected error for interface{} type, got nil")
		return
	}

	// Check that error is of correct type
	interfaceErr, ok := err.(*InterfaceTypeError)
	if !ok {
		t.Errorf("ValidateSchema() expected *InterfaceTypeError, got %T", err)
		return
	}

	if len(interfaceErr.Fields) != 1 {
		t.Errorf("Expected 1 interface field, got %d", len(interfaceErr.Fields))
	}

	// Check error message contains helpful information
	errMsg := err.Error()
	if !strings.Contains(errMsg, "interface{}") {
		t.Error("Error message should mention interface{}")
	}
	if !strings.Contains(errMsg, "data") {
		t.Error("Error message should mention the field name")
	}
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
	if err == nil {
		t.Error("ValidateSchema() expected error for []interface{} type, got nil")
		return
	}

	interfaceErr, ok := err.(*InterfaceTypeError)
	if !ok {
		t.Errorf("ValidateSchema() expected *InterfaceTypeError, got %T", err)
		return
	}

	if len(interfaceErr.Fields) != 1 {
		t.Errorf("Expected 1 interface field, got %d", len(interfaceErr.Fields))
	}

	if !interfaceErr.Fields[0].IsArray {
		t.Error("Expected IsArray to be true for []interface{}")
	}
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
	if err == nil {
		t.Error("ValidateSchema() expected error for slice with interface{} element, got nil")
		return
	}

	interfaceErr, ok := err.(*InterfaceTypeError)
	if !ok {
		t.Errorf("ValidateSchema() expected *InterfaceTypeError, got %T", err)
		return
	}

	if len(interfaceErr.Fields) != 1 {
		t.Errorf("Expected 1 interface field, got %d", len(interfaceErr.Fields))
	}

	if !interfaceErr.Fields[0].IsArray {
		t.Error("Expected IsArray to be true for slice with interface{} element")
	}
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
	if err == nil {
		t.Error("ValidateSchema() expected error for multiple interface{} types, got nil")
		return
	}

	interfaceErr, ok := err.(*InterfaceTypeError)
	if !ok {
		t.Errorf("ValidateSchema() expected *InterfaceTypeError, got %T", err)
		return
	}

	if len(interfaceErr.Fields) != 3 {
		t.Errorf("Expected 3 interface fields, got %d", len(interfaceErr.Fields))
	}
}

func TestValidateSchema_EmptySchema(t *testing.T) {
	s := &Schema{
		APIVersion: "example.com/v1",
		Kind:       "Example",
		Package:    "v1",
		Structs:    []StructDef{},
	}

	err := ValidateSchema(s)
	if err != nil {
		t.Errorf("ValidateSchema() expected no error for empty schema, got: %v", err)
	}
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
		if !strings.Contains(errMsg, required) {
			t.Errorf("Error message missing required string %q", required)
		}
	}
}
