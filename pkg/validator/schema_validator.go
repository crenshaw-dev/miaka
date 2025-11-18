package validator

import (
	"fmt"
	"strings"

	"github.com/crenshaw-dev/miaka/pkg/types"
)

// InterfaceTypeError represents an error caused by interface{} types in the schema
type InterfaceTypeError struct {
	Fields []InterfaceTypeLocation
}

// InterfaceTypeLocation describes where an interface{} type was found
type InterfaceTypeLocation struct {
	StructName string
	FieldName  string
	FieldPath  string
	YAMLPath   string
	Line       int
	IsArray    bool
}

func (e *InterfaceTypeError) Error() string {
	var sb strings.Builder

	sb.WriteString("❌ Cannot generate CRD: Found ")
	sb.WriteString(fmt.Sprintf("%d field(s)", len(e.Fields)))
	sb.WriteString(" with interface{} types\n\n")

	sb.WriteString("Kubernetes CRDs require concrete types and do not support interface{} (any type).\n")
	sb.WriteString("This usually happens when:\n")
	sb.WriteString("  • Empty arrays in YAML (e.g., `myField: []`)\n")
	sb.WriteString("  • Empty objects in YAML (e.g., `myField: {}`)\n\n")

	sb.WriteString("Problematic fields:\n")
	for i, field := range e.Fields {
		if i >= 5 {
			sb.WriteString(fmt.Sprintf("  ... and %d more\n", len(e.Fields)-5))
			break
		}
		if field.Line > 0 && field.YAMLPath != "" {
			sb.WriteString(fmt.Sprintf("  • Line %d: %s\n", field.Line, field.YAMLPath))
		} else {
			sb.WriteString(fmt.Sprintf("  • %s\n", field.FieldPath))
		}
		if field.IsArray {
			sb.WriteString("    (empty array - cannot infer element type)\n")
		}
	}

	sb.WriteString("\n💡 Solutions:\n")
	sb.WriteString("  1. Provide example values in your YAML:\n")
	sb.WriteString("     ```yaml\n")
	sb.WriteString("     # Instead of:\n")
	sb.WriteString("     imagePullSecrets: []\n")
	sb.WriteString("     \n")
	sb.WriteString("     # Use:\n")
	sb.WriteString("     imagePullSecrets:\n")
	sb.WriteString("     - name: my-secret\n")
	sb.WriteString("     ```\n\n")
	sb.WriteString("  2. Add type hints using Miaka markers:\n")
	sb.WriteString("     ```yaml\n")
	sb.WriteString("     # +miaka:type:map[string]string\n")
	sb.WriteString("     annotations: {}\n")
	sb.WriteString("     \n")
	sb.WriteString("     # +miaka:type:string (for array element type)\n")
	sb.WriteString("     imagePullSecrets: []\n")
	sb.WriteString("     ```\n\n")
	sb.WriteString("  3. For truly dynamic fields, consider using a different approach\n")
	sb.WriteString("     (Kubernetes CRDs are designed for structured schemas)\n")

	return sb.String()
}

// ValidateSchema checks for interface{} types in the schema before generation
func ValidateSchema(schema *types.Schema) error {
	var interfaceFields []InterfaceTypeLocation

	for _, structDef := range schema.Structs {
		for _, field := range structDef.Fields {
			isInterface := false
			isArray := false
			fieldPath := structDef.Name + "." + field.JSONName

			if field.IsSlice && field.ElemType == "interface{}" {
				isInterface = true
				isArray = true
			} else if field.Type == "interface{}" {
				isInterface = true
			} else if field.Type == "[]interface{}" {
				isInterface = true
				isArray = true
			}

			if isInterface {
				interfaceFields = append(interfaceFields, InterfaceTypeLocation{
					StructName: structDef.Name,
					FieldName:  field.Name,
					FieldPath:  fieldPath,
					YAMLPath:   field.YAMLPath,
					Line:       field.Line,
					IsArray:    isArray,
				})
			}
		}
	}

	if len(interfaceFields) > 0 {
		return &InterfaceTypeError{Fields: interfaceFields}
	}

	return nil
}
