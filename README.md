# Miaka

[![CI](https://github.com/crenshaw-dev/miaka/actions/workflows/ci.yaml/badge.svg)](https://github.com/crenshaw-dev/miaka/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/crenshaw-dev/miaka/branch/main/graph/badge.svg)](https://codecov.io/gh/crenshaw-dev/miaka)

> Make It A Kubernetes API

A golden chart is a Helm chart maintained by a platform team for use by developer teams. The values.yaml acts as the
developer's interface to the Internal Developer Platform, and the chart is the platform engineers' place to maintain
platform details.

To qualify as a golden chart, the Helm chart's values file must have an automatically generated OpenAPI spec and JSON schema as well as automatically generated documentation. The values file schema must adhere to the Kubernetes
Resource Model standard.

The values schema of a golden chart must be maintained with proper backwards compatibility and API versioning. Since
the chart's schema adheres to the KRM standard, the same best practices apply to golden chart schemas as apply to 
Kubernetes CRDs. It is recommended to follow Kubernetes CRD API versioning best practices and to take advantage of
automated tooling such as KAL to detect API design best practices.

A golden chart is meant to be rendered via a two-step call of `kustomize build`. The first call renders the golden
chart values file with any environment-specific overlays into an environment-specific values file. This is possible
because the values schema follows KRM.

The second `kustomize build` renders the final manifests by passing the environment-specific golden yaml as a values
file to the Helm chart inflator with the golden chart as a dependency. The second render will also include any 
Kustomize overlays, allowing last-mile tweaks that would not be possible via the values spec.

This two-step process may be done manually or via a miaka command that automates the process. The
CLI converts a simplified directory structure into the structure necessary for the kustomize commands.

```
base/
  kustomization.yaml # This contains a reference to the golden values openapi schema.
  values.yaml
environments/
  dev/
    kustomization.yaml
    values-patch.yaml
kustomization.yaml # This contains the helm chart inflator.
values-dev.yaml # This file is generated.
```

## How it works

miaka is a CLI with two main subcommands:

1. **init**: Converts a regular Helm values.yaml to KRM-compliant format
2. **build**: Generates Go types and CRD with OpenAPI v3 schema from KRM-compliant YAML

### Current Features

- **Values Initialization**: Convert existing Helm values.yaml to KRM format while preserving comments
- **Legacy Chart Support**: All fields remain at top level (not nested under spec) for minimal chart changes
- **YAML to Go Types**: Converts KRM-compliant YAML to Go struct definitions with proper types and tags
- **Type Inference**: Automatically infers Go types from YAML values (int, string, bool, arrays, objects)
- **Comment Preservation**: Maintains field descriptions and kubebuilder validation tags from YAML comments
- **CRD Generation**: Uses controller-gen library to generate Kubernetes CRD with OpenAPI v3 schema
- **JSON Schema Generation**: Extracts and converts OpenAPI schema to JSON Schema for Helm values validation
- **Dual Validation**: Validates input YAML against both CRD (Kubernetes) and JSON Schema (Helm)

## Usage

### Getting Help

```bash
# Show all available commands
miaka --help

# Get help for a specific command
miaka init --help
miaka build --help
```

### Shell Completion

The CLI includes shell completion support for bash, zsh, fish, and PowerShell:

```bash
# Generate completion script for your shell
miaka completion bash > /etc/bash_completion.d/miaka
miaka completion zsh > "${fpath[1]}/_miaka"
miaka completion fish > ~/.config/fish/completions/miaka.fish

# Or for the current session
source <(miaka completion bash)
```

### Init Command

Convert an existing Helm values.yaml to KRM-compliant format:

```bash
# Convert values.yaml to example.values.yaml
miaka init --api-version=myapp.io/v1 --kind=MyApp values.yaml

# With custom output file
miaka init --api-version=myapp.io/v1 --kind=MyApp -o custom.yaml values.yaml
```

**What it does:**
- Adds `apiVersion` and `kind` fields to your YAML
- Preserves all existing comments
- Keeps all fields at the top level (not nested under `spec`)
- The `metadata` field is added automatically by the build command as a Kubernetes implementation detail
- Ideal for legacy Helm charts where you want minimal changes

### Build Command

Generate types and CRDs from KRM-compliant YAML:

```bash
# Generate CRD and JSON Schema (default paths)
miaka build example.values.yaml

# Generate CRD and schema with custom paths
miaka build -c output/my-crd.yaml -s output/values.schema.json example.values.yaml

# Generate CRD, schema, and preserve types.go
miaka build -t types.go example.values.yaml

# Custom paths for all outputs
miaka build -t pkg/apis/v1/types.go -c crds/my-crd.yaml -s schemas/values.schema.json example.values.yaml
```

**What it does:**
- Generates CRD with OpenAPI v3 schema (default: `crd.yaml`)
- Generates JSON Schema for Helm validation (default: `values.schema.json`)
- Optionally preserves Go struct definitions with proper tags (use `-t` flag)
- Validates kubebuilder markers and field types
- Validates input YAML against both CRD and JSON Schema
- JSON Schema excludes `metadata` but includes `apiVersion`, `kind`, and all application fields

### Complete Workflow

```bash
# Step 1: Convert your existing values.yaml to KRM format
miaka init --api-version=myapp.io/v1 --kind=MyApp values.yaml

# Step 2: Edit example.values.yaml to add field descriptions and validation tags
# Add comments like:
#   # Number of replicas
#   # +kubebuilder:validation:Minimum=1
#   replicas: 3

# Step 3: Generate the CRD with OpenAPI schema
miaka build example.values.yaml

# The CRD now contains a full OpenAPI v3 schema with your descriptions and validations
```

### Future Features

- JSON schema generation from OpenAPI spec
- Documentation generation (README.md)
- Conversion between nested spec format and flat format

## example.values.yaml format

The example.values.yaml must follow the KRM spec. For example, it must have an apiVersion and a kind field.

Each field may have a description in the comment. It may also have comments with kubebuilder-style tags. The comment
is copied verbatim to the relevant go struct or field.

In lists of objects, we infer the fields of the object's struct from the example objects in the list. If there are
multiple items in the list, multiple objects may specify the same fields. But where fields are repeated, the comments
must be identical so that there is no ambiguity about how the field is to be documented.

Comments are only copied to go structs if they appear before the field in the yaml. Comments at the end of an example
line are discarded.
