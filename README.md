# Miaka

[![CI](https://github.com/crenshaw-dev/miaka/actions/workflows/ci.yaml/badge.svg)](https://github.com/crenshaw-dev/miaka/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/crenshaw-dev/miaka/branch/main/graph/badge.svg)](https://codecov.io/gh/crenshaw-dev/miaka)

> Make It A Kubernetes API

**Miaka** transforms your Helm chart values into fully-validated, Kubernetes-native APIs with automatic schema generation and breaking change detection.

## Why Miaka?

✨ **KRM Compliance** - Make your values files follow the Kubernetes Resource Model standard  
🚀 **Easy Schema Generation** - Automatically generate CRDs and JSON schemas from your YAML  
🛡️ **Breaking Change Detection** - Prevent API breakage with automatic compatibility checking

## Quick Start

### 1. Initialize your values file

Convert your existing Helm `values.yaml` to KRM format (or start from scratch):

```bash
miaka init --api-version=myapp.io/v1 --kind=MyApp values.yaml
```

This adds `apiVersion` and `kind` to your YAML while preserving all comments and keeping fields at the top level.

### 2. Generate your schemas

Build CRD and JSON Schema from your KRM-compliant YAML:

```bash
miaka build example.values.yaml
```

Generates:
- `crd.yaml` - Kubernetes CRD with OpenAPI v3 schema
- `values.schema.json` - JSON Schema for Helm validation

### 3. Update with confidence

Make changes to your values file and rebuild - miaka automatically detects breaking changes:

```bash
miaka build example.values.yaml
```

If you introduce breaking changes (like changing a field type), the build fails with clear error messages showing exactly what broke.

## Features

- **Comment-driven docs**: Add descriptions and kubebuilder validation tags as YAML comments
- **Type inference**: Automatically infers correct types from your example values
- **Dual validation**: Validates against both CRD (Kubernetes) and JSON Schema (Helm)
- **Legacy chart friendly**: Works with existing charts - no need to nest fields under `spec`

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

## Learn More

```bash
miaka --help
miaka init --help
miaka build --help
```

