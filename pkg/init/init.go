// Package init provides functionality for converting Helm values to KRM format.
package init

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// CheckKRMFields checks if a YAML file has apiVersion and kind fields
func CheckKRMFields(inputFile string) (hasApiVersion, hasKind bool) {
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return false, false
	}

	var rootNode yaml.Node
	if err := yaml.Unmarshal(data, &rootNode); err != nil {
		return false, false
	}

	if rootNode.Kind != yaml.DocumentNode || len(rootNode.Content) == 0 {
		return false, false
	}

	contentNode := rootNode.Content[0]
	if contentNode.Kind != yaml.MappingNode {
		return false, false
	}

	// Check for apiVersion and kind fields
	for i := 0; i < len(contentNode.Content); i += 2 {
		keyNode := contentNode.Content[i]
		if keyNode.Value == "apiVersion" {
			hasApiVersion = true
		}
		if keyNode.Value == "kind" {
			hasKind = true
		}
	}

	return hasApiVersion, hasKind
}

// ConvertToKRM converts a regular Helm values.yaml to a KRM-compliant YAML file
// by adding apiVersion and kind fields at the top level.
// All existing fields remain at the root level (not nested under spec).
// The metadata field is added automatically by the generator as an implementation detail.
//
// If inputFile is empty, creates an empty KRM-compliant YAML.
// If inputFile is provided but doesn't have apiVersion/kind, the provided apiVersion/kind are used.
// If inputFile already has apiVersion/kind, they are preserved and the provided values are ignored (can be empty).
// When apiVersion and kind are required but not provided, returns an error.
func ConvertToKRM(inputFile, outputFile, apiVersion, kind string) error {
	var contentNode *yaml.Node
	var rootNode yaml.Node

	// Handle empty input file (create new empty KRM YAML)
	if inputFile == "" {
		if apiVersion == "" || kind == "" {
			return fmt.Errorf("apiVersion and kind are required (provide via --api-version and --kind flags)")
		}

		// Create a new empty mapping node
		contentNode = &yaml.Node{
			Kind:    yaml.MappingNode,
			Content: []*yaml.Node{},
		}
		rootNode = yaml.Node{
			Kind:    yaml.DocumentNode,
			Content: []*yaml.Node{contentNode},
		}
	} else {
		// Read input file
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}

		// Parse YAML with full node structure to preserve comments
		if err := yaml.Unmarshal(data, &rootNode); err != nil {
			return fmt.Errorf("failed to parse YAML: %w", err)
		}

		// The root node is typically a DocumentNode, get the actual content node
		if rootNode.Kind != yaml.DocumentNode || len(rootNode.Content) == 0 {
			return fmt.Errorf("invalid YAML structure")
		}

		contentNode = rootNode.Content[0]

		// Check if this is a mapping node (object)
		if contentNode.Kind != yaml.MappingNode {
			return fmt.Errorf("root YAML node must be an object")
		}
	}

	// Check if apiVersion or kind already exist in the input
	var hasApiVersion, hasKind bool
	for i := 0; i < len(contentNode.Content); i += 2 {
		keyNode := contentNode.Content[i]
		if keyNode.Value == "apiVersion" {
			hasApiVersion = true
		}
		if keyNode.Value == "kind" {
			hasKind = true
		}
	}

	// Determine what values to use
	finalApiVersion := apiVersion
	finalKind := kind

	// Both exist in input - use those values (ignore provided flags)
	if !hasApiVersion || !hasKind {
		// Handle cases where one or both are missing
		if hasApiVersion {
			// Only apiVersion exists - missing kind
			return fmt.Errorf("file has apiVersion but missing kind - provide --kind flag or run interactively")
		}
		if hasKind {
			// Only kind exists - missing apiVersion
			return fmt.Errorf("file has kind but missing apiVersion - provide --api-version flag or run interactively")
		}
		// Neither exist, must provide both
		if finalApiVersion == "" || finalKind == "" {
			return fmt.Errorf("apiVersion and kind are required (provide via flags or run interactively)")
		}

		// Create new nodes for apiVersion and kind
		apiVersionKey := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "apiVersion",
		}
		apiVersionValue := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: finalApiVersion,
		}

		kindKey := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "kind",
		}
		kindValue := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: finalKind,
		}

		// Insert new nodes at the beginning while preserving existing nodes
		newContent := []*yaml.Node{
			apiVersionKey, apiVersionValue,
			kindKey, kindValue,
		}
		newContent = append(newContent, contentNode.Content...)
		contentNode.Content = newContent
	}

	// Marshal back to YAML
	output, err := yaml.Marshal(&rootNode)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Write to output file
	if err := os.WriteFile(outputFile, output, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
