# Miaka

[![CI](https://github.com/crenshaw-dev/miaka/actions/workflows/ci.yaml/badge.svg)](https://github.com/crenshaw-dev/miaka/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/crenshaw-dev/miaka/branch/main/graph/badge.svg)](https://codecov.io/gh/crenshaw-dev/miaka)

> Make It A Kubernetes API

**Miaka** transforms your Helm chart values into fully-validated, Kubernetes-native APIs with automatic schema generation and breaking change detection.

## Why Miaka?

* 🚀 **Easy Schema Generation** - Automatically generate CRDs and JSON schemas from your YAML
* 🛡️ **Breaking Change Detection** - Prevent API breakage with automatic compatibility checking
* ✨ **KRM Compliance** - Make your values files follow the Kubernetes Resource Model standard

## Installation

```bash
go install github.com/crenshaw-dev/miaka@latest
miaka version
```

## Quick Start

### 1. Initialize your values file

Convert your existing Helm `values.yaml` to KRM format (or start from scratch):

```bash
cd <path to your Helm chart>
miaka init
```

This generates:
- `example.values.yaml` - Complete example values file for generating schemas

### 2. Generate your schemas

Build CRD and JSON Schema from your KRM-compliant YAML:

```bash
miaka build
```

This automatically reads `example.values.yaml` and generates:
- `crd.yaml` - Kubernetes CRD with OpenAPI v3 schema
- `values.schema.json` - JSON Schema for Helm validation

### 3. Validate user values (optional)

Test that values files from your users pass the validation rules:

```bash
miaka validate user-values.yaml
```

This validates the values file against both your CRD and JSON Schema, helping you catch issues before deployment.

### 4. Update with confidence

Make changes to your values file and rebuild - miaka automatically detects breaking changes:

```bash
miaka build
```

If you introduce breaking changes (like changing a field type), the build fails with clear error messages showing exactly what broke.

Miaka uses crd.yaml to detect breaking changes, so make sure to keep that file!

## Features

- **Comment-driven docs**: Add descriptions and kubebuilder validation tags as YAML comments
- **Type inference**: Automatically infers correct types from your example values
- **Dual validation**: Validates against both CRD (Kubernetes) and JSON Schema (Helm)
- **Legacy chart friendly**: Works with existing charts - no need to change the structure

## Example

Add field documentation with comments:

```yaml
apiVersion: myapp.io/v1
kind: MyApp
# Number of replicas
# +kubebuilder:validation:Minimum=1
# +kubebuilder:validation:Maximum=10
replicas: 3
```

Miaka generates schemas with these descriptions and validations automatically.

## How It Works

Miaka doesn't reinvent the wheel - it brings together proven Kubernetes ecosystem tools:

1. **Schema Generation**: Uses [controller-gen](https://book.kubebuilder.io/reference/controller-gen.html) (the official Kubernetes CRD generator) to create OpenAPI v3 schemas from Go types
2. **Validation**: Leverages Helm's JSON Schema validation to verify your values against the generated schema
3. **Breaking Change Detection**: Employs [crdify](https://github.com/kubernetes-sigs/crdify) to catch API compatibility issues between versions

No magic, just battle-tested tools working together.

## Beyond Helm

While Miaka is designed with Helm charts in mind, nothing about it strictly requires Helm. At its core, Miaka helps you maintain a Kubernetes API based on a complete example file with validation markers. The generated CRD and JSON Schema can be used by any tool that processes YAML adhering to the API:

- **KRM Functions** - Process validated resources in Kustomize pipelines
- **Kubernetes Controllers** - Build operators that reconcile your custom resources

The Kubernetes Resource Model (KRM) format and OpenAPI v3 schemas are standards - any tool in the ecosystem can work with them.

Building a robust, validated API also means you can swap out the backend implementation (from Helm to a controller, or vice versa) with confidence that existing configurations will continue to work.

## Learn More

```bash
miaka --help
miaka init --help
miaka build --help
miaka validate --help
```

