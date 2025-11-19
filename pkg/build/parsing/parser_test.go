package parsing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/crenshaw-dev/miaka/pkg/build/schema"
)

const testKindName = "Example"

// TestNewParser tests parser initialization
func TestNewParser(t *testing.T) {
	p := NewParser()
	if p == nil {
		t.Fatal("NewParser returned nil")
	}
	if p.schema == nil {
		t.Error("Parser schema is nil")
	}
	if p.structNames == nil {
		t.Error("Parser structNames map is nil")
	}
	if len(p.schema.Structs) != 0 {
		t.Errorf("Expected empty structs, got %d", len(p.schema.Structs))
	}
}

// TestParseFile tests parsing from a file
func TestParseFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	yamlContent := `apiVersion: example.com/v1
kind: Example
replicas: 3
name: test
`
	if err := os.WriteFile(testFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	p := NewParser()
	s, err := p.ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if s.APIVersion != "example.com/v1" {
		t.Errorf("Expected apiVersion 'example.com/v1', got '%s'", s.APIVersion)
	}
	if s.Kind != testKindName {
		t.Errorf("Expected kind '%s', got '%s'", testKindName, s.Kind)
	}
	if s.Package != "v1" {
		t.Errorf("Expected package 'v1', got '%s'", s.Package)
	}
}

// TestParseFile_NonExistent tests error handling for non-existent files
func TestParseFile_NonExistent(t *testing.T) {
	p := NewParser()
	_, err := p.ParseFile("nonexistent.yaml")
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("Expected 'failed to read file' error, got: %v", err)
	}
}

// TestParse_BasicTypes tests parsing of basic scalar types
func TestParse_BasicTypes(t *testing.T) {
	yamlContent := `apiVersion: example.com/v1
kind: Example
stringField: "hello"
intField: 42
floatField: 3.14
boolField: true
`
	p := NewParser()
	s, err := p.Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(s.Structs) != 1 {
		t.Fatalf("Expected 1 struct, got %d", len(s.Structs))
	}

	mainStruct := s.Structs[0]
	if len(mainStruct.Fields) != 4 {
		t.Fatalf("Expected 4 fields, got %d", len(mainStruct.Fields))
	}

	// Check field types
	expectedTypes := map[string]string{
		"StringField": "string",
		"IntField":    "int",
		"FloatField":  "float64",
		"BoolField":   "bool",
	}

	for _, field := range mainStruct.Fields {
		expectedType, ok := expectedTypes[field.Name]
		if !ok {
			t.Errorf("Unexpected field: %s", field.Name)
			continue
		}
		if field.Type != expectedType {
			t.Errorf("Field %s: expected type %s, got %s", field.Name, expectedType, field.Type)
		}
	}
}

// TestParse_NestedObject tests parsing of nested objects
func TestParse_NestedObject(t *testing.T) {
	yamlContent := `apiVersion: example.com/v1
kind: Example
database:
  host: localhost
  port: 5432
  enabled: true
`
	p := NewParser()
	s, err := p.Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have main struct + nested Database struct
	if len(s.Structs) != 2 {
		t.Fatalf("Expected 2 structs, got %d", len(s.Structs))
	}

	// Find the DatabaseConfig struct
	var dbStruct *schema.StructDef
	for i := range s.Structs {
		if s.Structs[i].Name == "DatabaseConfig" {
			dbStruct = &s.Structs[i]
			break
		}
	}

	if dbStruct == nil {
		t.Fatal("DatabaseConfig struct not found")
	}

	if len(dbStruct.Fields) != 3 {
		t.Errorf("Expected 3 fields in Database struct, got %d", len(dbStruct.Fields))
	}

	// Verify field names and types
	expectedFields := map[string]string{
		"Host":    "string",
		"Port":    "int",
		"Enabled": "bool",
	}

	for _, field := range dbStruct.Fields {
		expectedType, ok := expectedFields[field.Name]
		if !ok {
			t.Errorf("Unexpected field in Database: %s", field.Name)
			continue
		}
		if field.Type != expectedType {
			t.Errorf("Field %s: expected type %s, got %s", field.Name, expectedType, field.Type)
		}
	}
}

// TestParse_Array tests parsing of arrays
func TestParse_Array(t *testing.T) {
	yamlContent := `apiVersion: example.com/v1
kind: Example
ports:
- 80
- 443
- 8080
tags:
- web
- api
`
	p := NewParser()
	s, err := p.Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	mainStruct := s.Structs[0]

	// Check ports field
	var portsField *schema.Field
	for i := range mainStruct.Fields {
		if mainStruct.Fields[i].Name == "Ports" {
			portsField = &mainStruct.Fields[i]
			break
		}
	}

	if portsField == nil {
		t.Fatal("Ports field not found")
	}

	if !portsField.IsSlice {
		t.Error("Ports field should be a slice")
	}
	if portsField.Type != "[]int" {
		t.Errorf("Expected Ports type '[]int', got '%s'", portsField.Type)
	}
	if portsField.ElemType != "int" {
		t.Errorf("Expected Ports elem type 'int', got '%s'", portsField.ElemType)
	}

	// Check tags field
	var tagsField *schema.Field
	for i := range mainStruct.Fields {
		if mainStruct.Fields[i].Name == "Tags" {
			tagsField = &mainStruct.Fields[i]
			break
		}
	}

	if tagsField == nil {
		t.Fatal("Tags field not found")
	}

	if !tagsField.IsSlice {
		t.Error("Tags field should be a slice")
	}
	if tagsField.Type != "[]string" {
		t.Errorf("Expected Tags type '[]string', got '%s'", tagsField.Type)
	}
}

// TestParse_ArrayOfObjects tests parsing arrays of objects
func TestParse_ArrayOfObjects(t *testing.T) {
	yamlContent := `apiVersion: example.com/v1
kind: Example
containers:
- name: web
  image: nginx:latest
  port: 80
- name: api
  image: node:16
  port: 3000
`
	p := NewParser()
	s, err := p.Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have main struct + Container struct
	if len(s.Structs) != 2 {
		t.Fatalf("Expected 2 structs, got %d", len(s.Structs))
	}

	// Find ContainersConfig struct
	var containerStruct *schema.StructDef
	for i := range s.Structs {
		if s.Structs[i].Name == "ContainersConfig" {
			containerStruct = &s.Structs[i]
			break
		}
	}

	if containerStruct == nil {
		t.Fatal("ContainersConfig struct not found")
	}

	if len(containerStruct.Fields) != 3 {
		t.Errorf("Expected 3 fields in Container, got %d", len(containerStruct.Fields))
	}

	// Verify containers field in Example (main) struct
	var mainStruct *schema.StructDef
	for i := range s.Structs {
		if s.Structs[i].Name == testKindName {
			mainStruct = &s.Structs[i]
			break
		}
	}

	if mainStruct == nil {
		t.Fatal("Example struct not found")
	}

	var containersField *schema.Field
	for i := range mainStruct.Fields {
		if mainStruct.Fields[i].Name == "Containers" {
			containersField = &mainStruct.Fields[i]
			break
		}
	}

	if containersField == nil {
		t.Fatal("Containers field not found")
	}

	if !containersField.IsSlice {
		t.Error("Containers should be a slice")
	}
	if containersField.Type != "[]ContainersConfig" {
		t.Errorf("Expected type '[]ContainersConfig', got '%s'", containersField.Type)
	}
}

// TestParse_EmptyArray tests parsing of empty arrays
func TestParse_EmptyArray(t *testing.T) {
	yamlContent := `apiVersion: example.com/v1
kind: Example
emptyList: []
`
	p := NewParser()
	s, err := p.Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	mainStruct := s.Structs[0]
	var emptyField *schema.Field
	for i := range mainStruct.Fields {
		if mainStruct.Fields[i].Name == "EmptyList" {
			emptyField = &mainStruct.Fields[i]
			break
		}
	}

	if emptyField == nil {
		t.Fatal("EmptyList field not found")
	}

	// Empty arrays without type hints should be []interface{}
	if emptyField.Type != "[]interface{}" {
		t.Errorf("Expected type '[]interface{}', got '%s'", emptyField.Type)
	}
}

// TestParse_TypeHints tests parsing with type hints
func TestParse_TypeHints(t *testing.T) {
	yamlContent := `apiVersion: example.com/v1
kind: Example
# +miaka:type:map[string]string
labels: {}
# +miaka:type:string
names: []
# +miaka:type:[]int
ports: []
`
	p := NewParser()
	s, err := p.Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	mainStruct := s.Structs[0]

	tests := []struct {
		fieldName    string
		expectedType string
	}{
		{"Labels", "map[string]string"},
		{"Names", "[]string"},
		{"Ports", "[]int"},
	}

	for _, tt := range tests {
		var field *schema.Field
		for i := range mainStruct.Fields {
			if mainStruct.Fields[i].Name == tt.fieldName {
				field = &mainStruct.Fields[i]
				break
			}
		}

		if field == nil {
			t.Errorf("Field %s not found", tt.fieldName)
			continue
		}

		if field.Type != tt.expectedType {
			t.Errorf("Field %s: expected type '%s', got '%s'", tt.fieldName, tt.expectedType, field.Type)
		}
	}
}

// TestParse_Comments tests comment extraction
func TestParse_Comments(t *testing.T) {
	yamlContent := `apiVersion: example.com/v1
kind: Example
# This is the replica count
# It must be a positive integer
replicas: 3
# Application name
name: myapp
`
	p := NewParser()
	s, err := p.Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	mainStruct := s.Structs[0]

	// Check replicas field comments
	var replicasField *schema.Field
	for i := range mainStruct.Fields {
		if mainStruct.Fields[i].Name == "Replicas" {
			replicasField = &mainStruct.Fields[i]
			break
		}
	}

	if replicasField == nil {
		t.Fatal("Replicas field not found")
	}

	if len(replicasField.Comments) != 2 {
		t.Errorf("Expected 2 comment lines for Replicas, got %d", len(replicasField.Comments))
	}

	if len(replicasField.Comments) >= 1 && !strings.Contains(replicasField.Comments[0], "replica count") {
		t.Errorf("Expected comment about replica count, got: %s", replicasField.Comments[0])
	}

	// Check name field comments
	var nameField *schema.Field
	for i := range mainStruct.Fields {
		if mainStruct.Fields[i].Name == "Name" {
			nameField = &mainStruct.Fields[i]
			break
		}
	}

	if nameField == nil {
		t.Fatal("Name field not found")
	}

	if len(nameField.Comments) != 1 {
		t.Errorf("Expected 1 comment line for Name, got %d", len(nameField.Comments))
	}
}

// TestParse_DeeplyNested tests deeply nested structures
func TestParse_DeeplyNested(t *testing.T) {
	yamlContent := `apiVersion: example.com/v1
kind: Example
config:
  database:
    connection:
      host: localhost
      port: 5432
    credentials:
      username: admin
      password: secret
`
	p := NewParser()
	s, err := p.Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have main + Config + DatabaseConfig + ConnectionConfig + CredentialsConfig = 5 structs
	if len(s.Structs) < 4 {
		t.Errorf("Expected at least 4 structs for deeply nested structure, got %d", len(s.Structs))
	}

	// Verify ConnectionConfig struct exists and has correct fields
	var connStruct *schema.StructDef
	for i := range s.Structs {
		if s.Structs[i].Name == "ConnectionConfig" {
			connStruct = &s.Structs[i]
			break
		}
	}

	if connStruct == nil {
		t.Fatal("ConnectionConfig struct not found")
	}

	if len(connStruct.Fields) != 2 {
		t.Errorf("Expected 2 fields in Connection, got %d", len(connStruct.Fields))
	}
}

// TestParse_StructNameCollision tests handling of struct name collisions
func TestParse_StructNameCollision(t *testing.T) {
	yamlContent := `apiVersion: example.com/v1
kind: Example
service:
  config:
    timeout: 30
database:
  config:
    timeout: 60
`
	p := NewParser()
	s, err := p.Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have unique struct names for all structs
	// Expected: Example, ServiceConfig, DatabaseConfig, and 2 nested config structs with unique names
	structNames := make(map[string]bool)

	for i := range s.Structs {
		name := s.Structs[i].Name
		if structNames[name] {
			t.Errorf("Duplicate struct name: %s", name)
		}
		structNames[name] = true
	}

	// Verify all names are unique (the test passes if no duplicates were found above)
	if len(structNames) != len(s.Structs) {
		t.Errorf("Expected all struct names to be unique, got %d unique names for %d structs", len(structNames), len(s.Structs))
	}
}

// TestParse_MetadataIgnored tests that metadata field is ignored
func TestParse_MetadataIgnored(t *testing.T) {
	yamlContent := `apiVersion: example.com/v1
kind: Example
metadata:
  name: test
  namespace: default
replicas: 3
`
	p := NewParser()
	s, err := p.Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	mainStruct := s.Structs[0]

	// Should only have replicas field, metadata should be ignored
	if len(mainStruct.Fields) != 1 {
		t.Errorf("Expected 1 field (metadata should be ignored), got %d", len(mainStruct.Fields))
	}

	if mainStruct.Fields[0].Name != "Replicas" {
		t.Errorf("Expected only Replicas field, got: %s", mainStruct.Fields[0].Name)
	}
}

// TestParse_InvalidYAML tests error handling for invalid YAML
func TestParse_InvalidYAML(t *testing.T) {
	invalidYAML := `this is not: valid: yaml: content`

	p := NewParser()
	_, err := p.Parse([]byte(invalidYAML))
	if err == nil {
		t.Fatal("Expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "failed to parse YAML") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

// TestParse_RootNotMapping tests error when root is not a mapping
func TestParse_RootNotMapping(t *testing.T) {
	yamlContent := `- item1
- item2
- item3
`
	p := NewParser()
	_, err := p.Parse([]byte(yamlContent))
	if err == nil {
		t.Fatal("Expected error when root is not a mapping")
	}
	if !strings.Contains(err.Error(), "root node must be a mapping") {
		t.Errorf("Expected 'root node must be a mapping' error, got: %v", err)
	}
}

// TestParse_MissingAPIVersion tests handling of missing apiVersion
func TestParse_MissingAPIVersion(t *testing.T) {
	yamlContent := `kind: Example
replicas: 3
`
	p := NewParser()
	s, err := p.Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if s.APIVersion != "" {
		t.Errorf("Expected empty apiVersion, got: %s", s.APIVersion)
	}
	if s.Kind != testKindName {
		t.Errorf("Expected kind '%s', got: %s", testKindName, s.Kind)
	}
}

// TestParse_YAMLPath tests that YAML paths are correctly set
func TestParse_YAMLPath(t *testing.T) {
	yamlContent := `apiVersion: example.com/v1
kind: Example
database:
  connection:
    host: localhost
`
	p := NewParser()
	s, err := p.Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find the ConnectionConfig struct
	var connStruct *schema.StructDef
	for i := range s.Structs {
		if s.Structs[i].Name == "ConnectionConfig" {
			connStruct = &s.Structs[i]
			break
		}
	}

	if connStruct == nil {
		t.Fatal("ConnectionConfig struct not found")
	}

	// Check that the host field has the correct YAML path
	var hostField *schema.Field
	for i := range connStruct.Fields {
		if connStruct.Fields[i].Name == "Host" {
			hostField = &connStruct.Fields[i]
			break
		}
	}

	if hostField == nil {
		t.Fatal("Host field not found")
	}

	expectedPath := "ConnectionConfig.host"
	if hostField.YAMLPath != expectedPath {
		t.Errorf("Expected YAMLPath '%s', got '%s'", expectedPath, hostField.YAMLPath)
	}
}

// TestParse_JSONNames tests that JSON names are preserved
func TestParse_JSONNames(t *testing.T) {
	yamlContent := `apiVersion: example.com/v1
kind: Example
camelCaseField: value1
snake_case_field: value2
kebab-case-field: value3
`
	p := NewParser()
	s, err := p.Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	mainStruct := s.Structs[0]

	tests := []struct {
		fieldName    string
		expectedJSON string
	}{
		{"CamelCaseField", "camelCaseField"},
		{"SnakeCaseField", "snake_case_field"},
		{"KebabCaseField", "kebab-case-field"},
	}

	for _, tt := range tests {
		var field *schema.Field
		for i := range mainStruct.Fields {
			if mainStruct.Fields[i].Name == tt.fieldName {
				field = &mainStruct.Fields[i]
				break
			}
		}

		if field == nil {
			t.Errorf("Field %s not found", tt.fieldName)
			continue
		}

		if field.JSONName != tt.expectedJSON {
			t.Errorf("Field %s: expected JSONName '%s', got '%s'", tt.fieldName, tt.expectedJSON, field.JSONName)
		}
	}
}

// TestExtractTypeHint tests the extractTypeHint function
func TestExtractTypeHint(t *testing.T) {
	tests := []struct {
		name     string
		comments []string
		expected string
	}{
		{
			name:     "simple type hint",
			comments: []string{"# +miaka:type:string"},
			expected: "string",
		},
		{
			name:     "type hint with space after colon",
			comments: []string{"# +miaka:type: map[string]string"},
			expected: "map[string]string",
		},
		{
			name:     "type hint among other comments",
			comments: []string{"# This is a field", "# +miaka:type:[]int", "# More comments"},
			expected: "[]int",
		},
		{
			name:     "no type hint",
			comments: []string{"# Just a regular comment"},
			expected: "",
		},
		{
			name:     "empty comments",
			comments: []string{},
			expected: "",
		},
		{
			name:     "type hint without hash",
			comments: []string{"+miaka:type:float64"},
			expected: "float64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTypeHint(tt.comments)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestParse_MergeListItems tests merging fields from multiple list items
func TestParse_MergeListItems(t *testing.T) {
	yamlContent := `apiVersion: example.com/v1
kind: Example
items:
- name: item1
  type: A
- name: item2
  type: B
  extra: value
`
	p := NewParser()
	s, err := p.Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find the ItemsConfig struct
	var itemStruct *schema.StructDef
	for i := range s.Structs {
		if s.Structs[i].Name == "ItemsConfig" {
			itemStruct = &s.Structs[i]
			break
		}
	}

	if itemStruct == nil {
		t.Fatal("ItemsConfig struct not found")
	}

	// Should have merged all fields: name, type, extra
	if len(itemStruct.Fields) != 3 {
		t.Errorf("Expected 3 merged fields, got %d", len(itemStruct.Fields))
	}

	fieldNames := make(map[string]bool)
	for _, field := range itemStruct.Fields {
		fieldNames[field.Name] = true
	}

	expectedFields := []string{"Name", "Type", "Extra"}
	for _, expectedField := range expectedFields {
		if !fieldNames[expectedField] {
			t.Errorf("Expected merged field '%s' not found", expectedField)
		}
	}
}

// TestParse_LineNumbers tests that line numbers are captured
func TestParse_LineNumbers(t *testing.T) {
	yamlContent := `apiVersion: example.com/v1
kind: Example
field1: value1
field2: value2
field3: value3
`
	p := NewParser()
	s, err := p.Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	mainStruct := s.Structs[0]

	for _, field := range mainStruct.Fields {
		if field.Line == 0 {
			t.Errorf("Field %s has line number 0, expected non-zero", field.Name)
		}
	}
}
