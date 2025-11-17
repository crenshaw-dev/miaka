# Golden Chart Generator

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

This two-step process may be done manually or via a golden-chart-generator command that automates the process. The
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

Golden chart generator is a CLI that accepts an example.values.yaml. The CLI generates a types.go in the format of
a kubebuilder CRD types.go. The CLI then calls controller-gen to produce an OpenAPI spec for the chart schema. 
Finally, the CLI converts the OpenAPI document to a a values.schema.json file.

The CLI optionally can accept a default values.yaml file for a chart and produce an example.values.yaml to start the
process.

The CLI also produces a README.md. The README contains a copy of the example.values.yaml with tag comments removed.
Only the field description is retained.

## example.values.yaml format

The example.values.yaml must follow the KRM spec. For example, it must have an apiVersion and a kind field.

Each field may have a description in the comment. It may also have comments with kubebuilder-style tags. The comment
is copied verbatim to the relevant go struct or field.

In lists of objects, we infer the fields of the object's struct from the example objects in the list. If there are
multiple items in the list, multiple objects may specify the same fields. But where fields are repeated, the comments
must be identical so that there is no ambiguity about how the field is to be documented.

Comments are only copied to go structs if they appear before the field in the yaml. Comments at the end of an example
line are discarded.
