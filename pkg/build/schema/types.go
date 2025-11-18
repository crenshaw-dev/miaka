package schema

// FieldType represents the type of a field in Go
type FieldType string

// Field type constants
const (
	TypeInt     FieldType = "int"
	TypeFloat64 FieldType = "float64"
	TypeString  FieldType = "string"
	TypeBool    FieldType = "bool"
)

// Field represents a single field in a struct
type Field struct {
	Name     string   // Go field name (PascalCase)
	JSONName string   // JSON tag name (original YAML name)
	Type     string   // Go type (can be primitive, struct name, or slice)
	Comments []string // Comment lines (including kubebuilder tags)
	IsSlice  bool     // Whether this is a slice type
	ElemType string   // Element type if IsSlice is true
	YAMLPath string   // Path in YAML (e.g., "global.imagePullSecrets")
	Line     int      // Line number in source YAML file
}

// StructDef represents a Go struct definition
type StructDef struct {
	Name     string   // Struct name (PascalCase)
	Comments []string // Comments for the struct
	Fields   []Field  // Fields in the struct
}

// Schema represents the complete parsed schema
type Schema struct {
	APIVersion string      // Kubernetes apiVersion
	Kind       string      // Kubernetes kind
	Package    string      // Go package name
	Structs    []StructDef // All struct definitions
}
