package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"strings"

	"github.com/crenshaw-dev/miaka/pkg/types"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Generator handles Go code generation using AST
type Generator struct {
	schema *types.Schema
	fset   *token.FileSet
}

// NewGenerator creates a new generator instance
func NewGenerator(schema *types.Schema) *Generator {
	return &Generator{
		schema: schema,
		fset:   token.NewFileSet(),
	}
}

// Generate generates Go code from the schema using AST
func (g *Generator) Generate() ([]byte, error) {
	// Create the AST file
	// Note: Package-level markers need to be in a separate doc.go file for proper controller-gen support
	file := &ast.File{
		Name: ast.NewIdent(g.schema.Package),
	}

	// Add imports
	file.Decls = append(file.Decls, g.generateImports())

	// Add main type (e.g., Example)
	file.Decls = append(file.Decls, g.generateMainType())

	// Generate all structs (except the main fields struct which was merged into the main type)
	for _, structDef := range g.schema.Structs {
		// Skip the struct that has the same name as Kind - its fields are on the main type
		if structDef.Name == g.schema.Kind {
			continue
		}
		file.Decls = append(file.Decls, g.generateStruct(structDef))
	}

	// Use go/printer to generate the code
	var buf bytes.Buffer
	cfg := printer.Config{
		Mode:     printer.UseSpaces | printer.TabIndent,
		Tabwidth: 8,
	}

	if err := cfg.Fprint(&buf, g.fset, file); err != nil {
		return nil, fmt.Errorf("failed to print AST: %w", err)
	}

	// Post-process to fix inline comments
	// The go/printer sometimes places field Doc comments inline when created programmatically
	// This fixes that by ensuring field comments appear on separate lines
	code := fixInlineComments(buf.String())

	// Use go/format for final cleanup
	formatted, err := format.Source([]byte(code))
	if err != nil {
		// Return unformatted code with error so user can debug
		return []byte(code), fmt.Errorf("failed to format code: %w", err)
	}

	return formatted, nil
}

// generateImports creates the import declaration
func (g *Generator) generateImports() *ast.GenDecl {
	return &ast.GenDecl{
		Tok:    token.IMPORT,
		Lparen: 1, // Force parentheses
		Specs: []ast.Spec{
			&ast.ImportSpec{
				Name: ast.NewIdent("metav1"),
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: `"k8s.io/apimachinery/pkg/apis/meta/v1"`,
				},
			},
		},
	}
}

// generateMainType generates the main KRM type (e.g., Example)
func (g *Generator) generateMainType() *ast.GenDecl {
	typeName := g.schema.Kind

	// Create doc comment with kubebuilder markers
	doc := &ast.CommentGroup{
		List: []*ast.Comment{
			{
				Text: "// +kubebuilder:object:root=true",
			},
			{
				Text: "//",
			},
			{
				Text: fmt.Sprintf("// %s is the Schema for the %ss API", typeName, strings.ToLower(typeName)),
			},
		},
	}

	// Start with KRM metadata fields
	fields := []*ast.Field{
		{
			// metav1.TypeMeta `json:",inline"`
			Type: &ast.SelectorExpr{
				X:   ast.NewIdent("metav1"),
				Sel: ast.NewIdent("TypeMeta"),
			},
			Tag: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "`json:\",inline\"`",
			},
		},
		{
			// metav1.ObjectMeta `json:"metadata,omitempty"`
			Type: &ast.SelectorExpr{
				X:   ast.NewIdent("metav1"),
				Sel: ast.NewIdent("ObjectMeta"),
			},
			Tag: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "`json:\"metadata,omitempty\"`",
			},
		},
	}

	// Find the main fields struct (same name as Kind)
	var mainFieldsDef *types.StructDef
	for i := range g.schema.Structs {
		if g.schema.Structs[i].Name == typeName {
			mainFieldsDef = &g.schema.Structs[i]
			break
		}
	}

	// Add all main fields directly to the type
	if mainFieldsDef != nil {
		for _, field := range mainFieldsDef.Fields {
			fields = append(fields, g.generateField(field))
		}
	}

	return &ast.GenDecl{
		Doc: doc,
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(typeName),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: fields,
					},
				},
			},
		},
	}
}

// generateStruct generates a struct definition
func (g *Generator) generateStruct(structDef types.StructDef) *ast.GenDecl {
	// Build doc comments
	var commentLines []string
	if len(structDef.Comments) > 0 {
		// Use first comment as description (lowercased)
		description := strings.ToLower(structDef.Comments[0])
		commentLines = append(commentLines, fmt.Sprintf("%s defines the %s", structDef.Name, description))
		// Add remaining comments
		commentLines = append(commentLines, structDef.Comments[1:]...)
	} else {
		// Auto-generate description
		description := g.generateStructDescription(structDef.Name)
		commentLines = append(commentLines, fmt.Sprintf("%s defines the %s", structDef.Name, description))
	}

	// Create comment group
	doc := g.createCommentGroup(commentLines)

	// Generate fields
	// Note: AST printer naturally adds spacing between fields with doc comments
	fields := make([]*ast.Field, 0, len(structDef.Fields))
	for _, field := range structDef.Fields {
		fields = append(fields, g.generateField(field))
	}

	return &ast.GenDecl{
		Doc: doc,
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(structDef.Name),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: fields,
					},
				},
			},
		},
	}
}

// generateField generates a field in a struct
func (g *Generator) generateField(field types.Field) *ast.Field {
	// Create doc comment for the field
	var doc *ast.CommentGroup
	if len(field.Comments) > 0 {
		doc = g.createCommentGroup(field.Comments)
	}

	// Determine field type
	var fieldType ast.Expr
	if field.IsSlice {
		fieldType = &ast.ArrayType{
			Elt: ast.NewIdent(field.ElemType),
		}
	} else {
		fieldType = ast.NewIdent(field.Type)
	}

	return &ast.Field{
		Doc:   doc,
		Names: []*ast.Ident{ast.NewIdent(field.Name)},
		Type:  fieldType,
		Tag: &ast.BasicLit{
			Kind:  token.STRING,
			Value: fmt.Sprintf("`json:\"%s,omitempty\"`", field.JSONName),
		},
	}
}

// createCommentGroup creates a comment group from comment lines
func (g *Generator) createCommentGroup(comments []string) *ast.CommentGroup {
	if len(comments) == 0 {
		return nil
	}

	commentList := make([]*ast.Comment, 0, len(comments))
	for i, comment := range comments {
		// Add slash position to force newline between comments
		slashPos := token.Pos(i + 1)
		commentList = append(commentList, &ast.Comment{
			Slash: slashPos,
			Text:  "// " + comment,
		})
	}

	return &ast.CommentGroup{
		List: commentList,
	}
}

// generateStructDescription creates a description for a struct
func (g *Generator) generateStructDescription(structName string) string {
	// Special handling for Config suffix
	if strings.HasSuffix(structName, "Config") {
		baseName := strings.TrimSuffix(structName, "Config")
		words := splitPascalCase(baseName)
		description := strings.ToLower(strings.Join(words, " "))
		return description + " configuration"
	}

	// Convert PascalCase to words
	words := splitPascalCase(structName)
	description := strings.ToLower(strings.Join(words, " "))
	return description
}

// splitPascalCase splits a PascalCase string into words
func splitPascalCase(s string) []string {
	var words []string
	var currentWord strings.Builder

	for i, r := range s {
		if i > 0 && isUpper(r) && (i+1 < len(s) && !isUpper(rune(s[i+1])) || !isUpper(rune(s[i-1]))) {
			words = append(words, currentWord.String())
			currentWord.Reset()
		}
		currentWord.WriteRune(r)
	}

	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

// isUpper checks if a rune is uppercase
func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

// extractGroup extracts the group from an apiVersion using Kubernetes libraries
// e.g., "example.com/v1alpha1" -> "example.com"
func extractGroup(apiVersion string) string {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		// Fallback to original string if parsing fails
		return apiVersion
	}
	return gv.Group
}

// fixInlineComments fixes comments that appear inline with struct/field declarations
// This is a workaround for go/printer placing Doc comments inline when ASTs are created programmatically
func fixInlineComments(code string) string {
	lines := strings.Split(code, "\n")
	var result []string

	i := 0
	for i < len(lines) {
		line := lines[i]

		// Handle struct declaration with inline comment (e.g., "type Foo struct {// comment" or "struct { // comment")
		if strings.Contains(line, "struct {") && strings.Contains(line, "//") {
			// Try both patterns: "{//" and "{ //"
			var parts []string
			if strings.Contains(line, "{// ") {
				parts = strings.SplitN(line, "{// ", 2)
			} else if strings.Contains(line, "{ // ") {
				parts = strings.SplitN(line, "{ // ", 2)
			}

			if len(parts) == 2 {
				indent := getIndent(line)
				// Add struct line without comment
				result = append(result, parts[0]+"{")
				// Add comment on next line with field indentation
				result = append(result, indent+"\t"+"// "+parts[1])
				i++
				continue
			}
		}

		// Handle field with inline comment (e.g., "Field type `json:...`// comment")
		// The comment actually belongs to the NEXT field, not this one
		if strings.Contains(line, "`json:") && strings.Contains(line, "`//") {
			parts := strings.SplitN(line, "`// ", 2)
			if len(parts) == 2 {
				indent := getIndent(line)
				comment := parts[1]

				// Add current field without comment
				result = append(result, parts[0]+"`")

				// Add blank line after field
				result = append(result, "")

				// Add the comment before the NEXT field (if there is one)
				if i+1 < len(lines) {
					nextLine := strings.TrimSpace(lines[i+1])
					if nextLine != "}" && nextLine != "" {
						result = append(result, indent+"// "+comment)
					}
				}

				i++
				continue
			}
		}

		result = append(result, line)
		i++
	}

	return strings.Join(result, "\n")
}

// getIndent returns the leading whitespace/tabs from a line
func getIndent(line string) string {
	indent := ""
	for _, r := range line {
		if r == '\t' || r == ' ' {
			indent += string(r)
		} else {
			break
		}
	}
	return indent
}
