# Test Data

This directory contains test cases for validating the Miaka tool's output.

## Structure

Each subdirectory represents a test case with the following files:

```
testdata/
  <test-case-name>/
    input.yaml           - KRM-compliant input YAML
    expected_types.go    - Expected Go types output
    expected_crd.yaml    - Expected CRD output
    .skip                - (Optional) Mark test as skipped
```

## Automated Testing

The `integration_test.go` file at the project root automatically discovers and runs all test cases in this directory.

Run all integration tests:
```bash
make test
# or specifically:
go test -v -run TestDataTestCases
```

Each test case automatically:
1. Generates Go types from `input.yaml`
2. Compares output with `expected_types.go`
3. Generates CRD from the types
4. Compares output with `expected_crd.yaml`

## Skipping Test Cases

To skip a test case (e.g., for known limitations), create a `.skip` file in the test case directory:

```bash
echo "Reason for skipping this test" > testdata/my-case/.skip
```

The test will be marked as skipped and won't cause test failures.

## Test Cases

### `basic/`
A comprehensive example demonstrating:
- Multiple field types (int, string, bool)
- Nested structs (ServiceConfig)
- Arrays (EnvConfig)
- Kubebuilder validation markers

### `minimal/`
A minimal example with just a single field to test basic functionality.

### `argo-events/` (skipped)
A complex real-world example that exposes current limitations (interface{} types not supported).

## Adding New Test Cases

1. Create a new directory under `testdata/`:
   ```bash
   mkdir testdata/my-new-case
   ```

2. Add your `input.yaml` file with KRM-compliant YAML

3. Generate expected outputs:
   ```bash
   miaka build testdata/my-new-case/input.yaml \
     -o testdata/my-new-case/expected_types.go \
     -c testdata/my-new-case/expected_crd.yaml \
     --keep-types
   ```

4. Review the generated files to ensure they match expectations

5. Run tests to verify:
   ```bash
   make test
   ```

## Notes

- `go.mod` and `go.sum` are intermediate files created during CRD generation
- These are created in temporary directories and automatically cleaned up
- They are not part of the expected test outputs
- Controller-gen version annotations in CRDs are normalized during comparison to avoid false failures

