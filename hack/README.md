# Maintenance Scripts

## `update-apimachinery.sh`

Updates `k8s.io/apimachinery` to the latest version and regenerates the embedded `go.mod` and `go.sum` files used for CRD generation.

### Usage

```bash
./hack/update-apimachinery.sh
```

### What it does

1. Fetches the latest version of `k8s.io/apimachinery`
2. Generates a complete `go.mod` with all transitive dependencies
3. Generates the corresponding `go.sum` with checksums
4. Copies both files to `pkg/build/generation/crd/embedded/`

### After running

1. Run tests to ensure compatibility:
   ```bash
   go test ./pkg/build/generation/crd -v
   ```

2. Commit the changes:
   ```bash
   git add pkg/build/generation/crd/embedded/
   git commit -m "Update k8s.io/apimachinery to <version>"
   ```

### Why this is needed

Miaka embeds static `go.mod` and `go.sum` files to avoid requiring the `go` toolchain at runtime. This makes the binary truly standalone and eliminates the overhead of running `go mod tidy` during CRD generation. However, it means we need to regenerate these files when updating `k8s.io/apimachinery`.

