// Package validation provides validation utilities for CRDs and YAML files.
package validation

import (
	"fmt"
	"os"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/crdify/pkg/config"
	"sigs.k8s.io/crdify/pkg/runner"
	"sigs.k8s.io/crdify/pkg/validations"
)

// CheckBreakingChanges compares an existing CRD file with a newly generated CRD
// and returns an error if breaking changes are detected.
// If the old CRD file doesn't exist, no error is returned (first-time generation).
func CheckBreakingChanges(oldCRDPath string, newCRDContent []byte) error {
	// Check if old CRD exists
	if _, err := os.Stat(oldCRDPath); os.IsNotExist(err) {
		// No existing CRD, skip validation
		return nil
	}

	// Load old CRD
	oldCRD, err := loadCRDFromFile(oldCRDPath)
	if err != nil {
		return fmt.Errorf("failed to load existing CRD from %s: %w", oldCRDPath, err)
	}

	// Parse new CRD from content
	newCRD := &apiextensionsv1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(newCRDContent, newCRD); err != nil {
		return fmt.Errorf("failed to unmarshal new CRD: %w", err)
	}

	// Create default config for crdify
	cfg := &config.Config{
		UnhandledEnforcement: config.EnforcementPolicyNone,
		Conversion:           config.ConversionPolicyNone,
	}

	// Create runner with default validations
	r, err := runner.New(cfg, runner.DefaultRegistry())
	if err != nil {
		return fmt.Errorf("failed to create crdify runner: %w", err)
	}

	// Run validations
	results := r.Run(oldCRD, newCRD)

	// Check for breaking changes (errors)
	if results.HasFailures() {
		// Format the results as plain text for error message
		output := renderErrorsOnly(results)
		return fmt.Errorf("breaking changes detected:\n%s", output)
	}

	return nil
}

// loadCRDFromFile loads a CRD from a file path
func loadCRDFromFile(filePath string) (*apiextensionsv1.CustomResourceDefinition, error) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	crd := &apiextensionsv1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(fileBytes, crd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CRD: %w", err)
	}

	return crd, nil
}

// renderErrorsOnly renders only the validation errors, not all passing checks
// Based on crdify's RenderPlainText but filtered to errors only
func renderErrorsOnly(results *runner.Results) string {
	var out strings.Builder

	// CRD Validations
	renderCRDValidationErrors(&out, results.CRDValidation)

	// Same Version Validations
	renderVersionValidationErrors(&out, results.SameVersionValidation)

	// Served Version Validations
	renderVersionValidationErrors(&out, results.ServedVersionValidation)

	return out.String()
}

// renderCRDValidationErrors renders CRD validation errors
func renderCRDValidationErrors(out *strings.Builder, validationResults []validations.ComparisonResult) {
	for _, result := range validationResults {
		if len(result.Errors) > 0 {
			for _, err := range result.Errors {
				fmt.Fprintf(out, "- %s - %s\n", result.Name, err)
			}
		}
	}
}

// renderVersionValidationErrors renders version validation errors
func renderVersionValidationErrors(out *strings.Builder, validationResults map[string]map[string][]validations.ComparisonResult) {
	for version, versionResults := range validationResults {
		for property, propertyResults := range versionResults {
			for _, propertyResult := range propertyResults {
				if len(propertyResult.Errors) > 0 {
					for _, err := range propertyResult.Errors {
						fmt.Fprintf(out, "- %s - %s - %s - %s\n", version, property, propertyResult.Name, err)
					}
				}
			}
		}
	}
}
