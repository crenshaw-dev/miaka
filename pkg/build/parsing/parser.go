// Package parsing provides YAML parsing functionality with comment preservation.
package parsing

import (
	"fmt"
	"os"
	"strings"

	"github.com/crenshaw-dev/miaka/pkg/build/schema"
	"gopkg.in/yaml.v3"
)

// Parser handles YAML parsing with comment preservation
type Parser struct {
	schema      *schema.Schema
	structNames map[string]bool // Track used struct names to avoid collisions
}

// NewParser creates a new parser instance
func NewParser() *Parser {
	return &Parser{
		schema: &schema.Schema{
			Structs: make([]schema.StructDef, 0),
		},
		structNames: make(map[string]bool),
	}
}

// ParseFile parses a YAML file and returns a Schema
func (p *Parser) ParseFile(filename string) (*schema.Schema, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return p.Parse(data)
}

// Parse parses YAML data and returns a Schema
func (p *Parser) Parse(data []byte) (*schema.Schema, error) {
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// The root should be a document node
	if node.Kind != yaml.DocumentNode || len(node.Content) == 0 {
		return nil, fmt.Errorf("invalid YAML structure")
	}

	rootMap := node.Content[0]
	if rootMap.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("root node must be a mapping")
	}

	// Parse top-level fields
	if err := p.parseRootNode(rootMap); err != nil {
		return nil, err
	}

	return p.schema, nil
}

// parseRootNode parses the root mapping node
func (p *Parser) parseRootNode(node *yaml.Node) error {
	// First pass: collect apiVersion and kind
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		key := keyNode.Value

		switch key {
		case "apiVersion":
			p.schema.APIVersion = valueNode.Value
			version, err := schema.ParseAPIVersion(valueNode.Value)
			if err != nil {
				return err
			}
			p.schema.Package = version
		case "kind":
			p.schema.Kind = valueNode.Value
		}
	}

	// Second pass: parse all top-level fields (except apiVersion, kind, metadata)
	// These become fields directly on the main type
	mainFields := &schema.StructDef{
		Name:     p.schema.Kind,
		Fields:   []schema.Field{},
		Comments: []string{},
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		key := keyNode.Value

		// Skip KRM metadata fields
		if key == "apiVersion" || key == "kind" || key == "metadata" {
			continue
		}

		// Parse this field - it will be added directly to the main type
		comments := extractComments(keyNode)
		field, nestedStructs, err := p.parseFieldWithPath(key, key, valueNode, comments)
		if err != nil {
			return fmt.Errorf("failed to parse field %s: %w", key, err)
		}
		mainFields.Fields = append(mainFields.Fields, *field)

		// Add any nested structs to the schema
		p.schema.Structs = append(p.schema.Structs, nestedStructs...)
	}

	// Add the main fields struct (this will be merged into the main type by the generator)
	p.schema.Structs = append(p.schema.Structs, *mainFields)

	return nil
}

// parseObject parses a mapping node into a struct definition
func (p *Parser) parseObject(node *yaml.Node, structName string, structComments []string) (*schema.StructDef, error) {
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected mapping node")
	}

	structDef := &schema.StructDef{
		Name:     structName,
		Comments: structComments,
		Fields:   make([]schema.Field, 0),
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		fieldName := keyNode.Value
		comments := extractComments(keyNode)

		// Build the yaml path for nested structs
		yamlPath := fmt.Sprintf("%s.%s", structName, fieldName)
		field, nestedStructs, err := p.parseFieldWithPath(fieldName, yamlPath, valueNode, comments)
		if err != nil {
			return nil, fmt.Errorf("failed to parse field %s: %w", fieldName, err)
		}

		structDef.Fields = append(structDef.Fields, *field)

		// Add nested structs to schema
		p.schema.Structs = append(p.schema.Structs, nestedStructs...)
	}

	return structDef, nil
}

// parseFieldWithPath parses a field with YAML path tracking
func (p *Parser) parseFieldWithPath(fieldName string, yamlPath string, valueNode *yaml.Node, comments []string) (*schema.Field, []schema.StructDef, error) {
	field := &schema.Field{
		Name:     schema.ToPascalCase(fieldName),
		JSONName: fieldName,
		Comments: schema.FormatComments(comments),
		YAMLPath: yamlPath,
		Line:     valueNode.Line,
	}

	nestedStructs := make([]schema.StructDef, 0)

	// Check for explicit type hint in comments
	typeHint := extractTypeHint(comments)

	switch valueNode.Kind {
	case yaml.ScalarNode:
		// Infer type from the scalar value
		var value interface{}
		if err := valueNode.Decode(&value); err != nil {
			return nil, nil, fmt.Errorf("failed to decode scalar: %w", err)
		}
		field.Type = schema.InferType(value)

	case yaml.MappingNode:
		// This is a nested object
		if len(valueNode.Content) == 0 && typeHint != "" {
			// Empty object with type hint (e.g., +miaka:type:map[string]string)
			field.Type = typeHint
		} else {
			// Non-empty object or no type hint
			structName := p.generateUniqueStructName(fieldName, yamlPath)
			field.Type = structName

			structComments := extractCommentsForStruct(valueNode)
			nestedStruct, err := p.parseObject(valueNode, structName, structComments)
			if err != nil {
				return nil, nil, err
			}
			nestedStructs = append(nestedStructs, *nestedStruct)
		}

	case yaml.SequenceNode:
		// This is a list
		field.IsSlice = true

		if len(valueNode.Content) == 0 {
			// Empty list - check for type hint first
			if typeHint != "" {
				// Type hint provided (e.g., +miaka:type:[]string or +miaka:type:string for array elements)
				if strings.HasPrefix(typeHint, "[]") {
					field.Type = typeHint
					field.ElemType = strings.TrimPrefix(typeHint, "[]")
				} else {
					// Assume the hint is for the element type
					field.Type = "[]" + typeHint
					field.ElemType = typeHint
				}
			} else {
				// No type hint, can't infer type
				field.Type = "[]interface{}"
				field.ElemType = "interface{}"
			}
		} else {
			// Examine the first element to determine type
			firstElem := valueNode.Content[0]

			// Check for struct-level comments (between list colon and first dash)
			structComments := extractListItemComments(valueNode)

			switch firstElem.Kind {
			case yaml.ScalarNode:
				var value interface{}
				if err := firstElem.Decode(&value); err != nil {
					return nil, nil, fmt.Errorf("failed to decode list element: %w", err)
				}
				elemType := schema.InferType(value)
				field.ElemType = elemType
				field.Type = "[]" + elemType

			case yaml.MappingNode:
				// List of objects - need to merge fields from all elements
				structName := p.generateUniqueStructName(fieldName, yamlPath)
				field.ElemType = structName
				field.Type = "[]" + structName

				// Merge all fields from all list items
				mergedStruct, err := p.mergeListItems(valueNode, structName, structComments)
				if err != nil {
					return nil, nil, err
				}
				nestedStructs = append(nestedStructs, *mergedStruct)
			}
		}
	}

	return field, nestedStructs, nil
}

// generateUniqueStructName creates a unique struct name, adding prefixes if there's a collision
func (p *Parser) generateUniqueStructName(fieldName string, yamlPath string) string {
	baseName := schema.GenerateStructName(fieldName)

	// If no collision, use the base name
	if !p.structNames[baseName] {
		p.structNames[baseName] = true
		return baseName
	}

	// Collision detected - use the parent path to make it unique
	// Extract parent from yamlPath (e.g., "configs.nats.versions" -> "Nats")
	pathParts := strings.Split(yamlPath, ".")
	if len(pathParts) > 1 {
		// Use the parent field name as prefix
		parentName := schema.ToPascalCase(pathParts[len(pathParts)-2])
		uniqueName := parentName + baseName

		// If still collision, add more context
		counter := 2
		originalUniqueName := uniqueName
		for p.structNames[uniqueName] {
			uniqueName = fmt.Sprintf("%s%d", originalUniqueName, counter)
			counter++
		}

		p.structNames[uniqueName] = true
		return uniqueName
	}

	// Fallback: add numeric suffix
	counter := 2
	uniqueName := fmt.Sprintf("%s%d", baseName, counter)
	for p.structNames[uniqueName] {
		counter++
		uniqueName = fmt.Sprintf("%s%d", baseName, counter)
	}

	p.structNames[uniqueName] = true
	return uniqueName
}

// mergeListItems merges fields from all items in a list to create a single struct
func (p *Parser) mergeListItems(sequenceNode *yaml.Node, structName string, structComments []string) (*schema.StructDef, error) {
	mergedFields := make(map[string]*schema.Field)
	var fieldOrder []string

	for _, itemNode := range sequenceNode.Content {
		if itemNode.Kind != yaml.MappingNode {
			continue
		}

		for i := 0; i < len(itemNode.Content); i += 2 {
			keyNode := itemNode.Content[i]
			valueNode := itemNode.Content[i+1]

			fieldName := keyNode.Value
			comments := extractComments(keyNode)

			if existingField, exists := mergedFields[fieldName]; exists {
				// Field already exists, verify comments match
				existingComments := strings.Join(existingField.Comments, "\n")
				newComments := strings.Join(schema.FormatComments(comments), "\n")
				if existingComments != newComments && len(comments) > 0 {
					return nil, fmt.Errorf("conflicting comments for field %s in list items", fieldName)
				}
			} else {
				// New field - build yaml path
				yamlPath := fmt.Sprintf("%s.%s", structName, fieldName)
				field, nestedStructs, err := p.parseFieldWithPath(fieldName, yamlPath, valueNode, comments)
				if err != nil {
					return nil, err
				}
				mergedFields[fieldName] = field
				fieldOrder = append(fieldOrder, fieldName)

				// Add nested structs
				p.schema.Structs = append(p.schema.Structs, nestedStructs...)
			}
		}
	}

	// Build the struct with fields in order
	structDef := &schema.StructDef{
		Name:     structName,
		Comments: structComments,
		Fields:   make([]schema.Field, 0, len(mergedFields)),
	}

	for _, fieldName := range fieldOrder {
		structDef.Fields = append(structDef.Fields, *mergedFields[fieldName])
	}

	return structDef, nil
}

// extractTypeHint looks for +miaka:type:<type> marker in comments
func extractTypeHint(comments []string) string {
	for _, comment := range comments {
		trimmed := strings.TrimSpace(comment)
		// Remove leading # if present
		trimmed = strings.TrimPrefix(trimmed, "#")
		trimmed = strings.TrimSpace(trimmed)

		// Check for +miaka:type: marker (with or without space after colon)
		if strings.HasPrefix(trimmed, "+miaka:type:") {
			typeStr := strings.TrimPrefix(trimmed, "+miaka:type:")
			typeStr = strings.TrimSpace(typeStr) // Allow space after colon
			if typeStr != "" {
				return typeStr
			}
		}
	}
	return ""
}

// extractComments extracts head comments from a node
func extractComments(node *yaml.Node) []string {
	comments := make([]string, 0)
	if node.HeadComment != "" {
		lines := strings.Split(node.HeadComment, "\n")
		comments = append(comments, lines...)
	}
	return comments
}

// extractCommentsForStruct extracts comments that should apply to a struct
// For objects, we use head comments on the mapping node
func extractCommentsForStruct(_ *yaml.Node) []string {
	// For structs defined by objects, no additional comments beyond field comments
	return []string{}
}

// extractListItemComments extracts comments for list item structs
// These are comments between the list colon and the first dash
func extractListItemComments(sequenceNode *yaml.Node) []string {
	comments := make([]string, 0)

	// Look at the first item's head comment
	if len(sequenceNode.Content) > 0 {
		firstItem := sequenceNode.Content[0]
		if firstItem.HeadComment != "" {
			lines := strings.Split(firstItem.HeadComment, "\n")
			comments = append(comments, lines...)
		}
	}

	return schema.FormatComments(comments)
}
