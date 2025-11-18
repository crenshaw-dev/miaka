// Package schema provides types and utilities for representing data schemas.
package schema

import (
	"fmt"
	"strings"
	"unicode"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// InferType determines the Go type from a YAML value
func InferType(value interface{}) string {
	switch v := value.(type) {
	case int, int64:
		return "int"
	case float64:
		// Check if it's actually an integer
		if v == float64(int(v)) {
			return "int"
		}
		return "float64"
	case string:
		return "string"
	case bool:
		return "bool"
	default:
		return "interface{}"
	}
}

// ToPascalCase converts a camelCase or snake_case string to PascalCase
// Also sanitizes invalid Go identifier characters (., /, :, etc.)
func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}

	// Split on common separators: _, -, ., /, :
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == '.' || r == '/' || r == ':'
	})

	if len(parts) == 0 {
		// Shouldn't happen but handle gracefully
		return sanitizeIdentifier(s)
	}

	// Convert each part to title case
	var result strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
			result.WriteString(string(runes))
		}
	}

	resultStr := result.String()
	if resultStr == "" {
		// Fallback if all parts were empty
		return sanitizeIdentifier(s)
	}

	return resultStr
}

// sanitizeIdentifier removes or replaces invalid Go identifier characters
func sanitizeIdentifier(s string) string {
	// Replace common invalid chars with nothing
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "-", "")

	// Capitalize first letter
	runes := []rune(s)
	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
	}

	return string(runes)
}

// GenerateStructName creates a struct name from a field name
// For example: "service" -> "ServiceConfig", "env" -> "EnvConfig"
func GenerateStructName(fieldName string) string {
	base := ToPascalCase(fieldName)

	// Add "Config" suffix if not already present
	if !strings.HasSuffix(base, "Config") && !strings.HasSuffix(base, "Spec") {
		return base + "Config"
	}

	return base
}

// FormatComments formats comment lines for Go documentation
func FormatComments(comments []string) []string {
	var result []string
	for _, comment := range comments {
		// Trim whitespace
		comment = strings.TrimSpace(comment)
		// Remove leading # if present
		comment = strings.TrimPrefix(comment, "#")
		comment = strings.TrimSpace(comment)
		if comment != "" {
			result = append(result, comment)
		}
	}
	return result
}

// ParseAPIVersion extracts the version from an apiVersion string using Kubernetes libraries
// e.g., "example.com/v1alpha1" -> "v1alpha1"
func ParseAPIVersion(apiVersion string) (string, error) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return "", fmt.Errorf("invalid apiVersion format: %s: %w", apiVersion, err)
	}
	return gv.Version, nil
}
