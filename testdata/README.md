# Test Data Structure

This directory contains test cases for the `miaka` CLI commands.

## Directory Structure

```
testdata/
├── build/          # Test cases for 'miaka build' command
│   ├── basic/
│   ├── minimal/
│   └── argo-events/
└── init/           # Test cases for 'miaka init' command
    ├── basic-conversion/
    ├── empty-file/
    ├── nested-structure/
    ├── comments-preservation/
    ├── array-handling/
    └── existing-krm/
```

## Build Test Cases

Each test case directory under `testdata/build/` should contain:

- **`input.yaml`** (required) - The input YAML file to process
- **`expected_types.go`** (optional) - Expected generated Go types
- **`expected_crd.yaml`** (optional) - Expected generated CRD
- **`expected_schema.json`** (optional) - Expected generated JSON Schema
- **`.skip`** (optional) - If present, the test will be skipped. The file contents will be used as the skip reason.

The test will:
1. Run `miaka build` on `input.yaml`
2. Generate types, CRD, and JSON Schema
3. Compare outputs with expected files (if they exist)

## Init Test Cases

Each test case directory under `testdata/init/` should contain:

- **`input.yaml`** (optional) - The input values.yaml file. If absent, tests creating empty KRM files.
- **`expected.yaml`** (required) - The expected output example.values.yaml
- **`flags.txt`** (optional) - Command-line flags to pass, one per line (e.g., `--api-version=myapp.io/v1`)
- **`.skip`** (optional) - If present, the test will be skipped. The file contents will be used as the skip reason.

The test will:
1. Run `miaka init` with the specified flags
2. Convert `input.yaml` (or create empty file if no input)
3. Compare output with `expected.yaml`

## Adding New Test Cases

To add a new test case:

1. Create a new directory under `testdata/build/` or `testdata/init/`
2. Add the required files (see above)
3. Run the tests to verify

The test framework will automatically discover and run all test cases in these directories.

## Skipping Tests

To skip a test temporarily:

```bash
echo "Reason for skipping" > testdata/build/my-test/.skip
```

The test will be skipped and the reason will be displayed in the test output.
