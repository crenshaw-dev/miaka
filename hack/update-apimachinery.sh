#!/bin/bash
set -euo pipefail

# Script to update k8s.io/apimachinery to the latest version
# This regenerates the embedded go.mod and go.sum files used by CRD generation

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
EMBEDDED_DIR="$ROOT_DIR/pkg/build/generation/crd/embedded"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Updating k8s.io/apimachinery to latest version...${NC}"

# Create temporary directory
WORK_DIR=$(mktemp -d)
trap "rm -rf $WORK_DIR" EXIT

cd "$WORK_DIR"

# Get current Go version (just major.minor)
GO_VERSION=$(go version | sed -E 's/.*go([0-9]+\.[0-9]+).*/\1/')
echo "Using Go version: $GO_VERSION"

# Initialize module with latest k8s.io/apimachinery
echo "Fetching latest k8s.io/apimachinery..."
cat > go.mod << EOF
module test

go $GO_VERSION
EOF

# Create a minimal types file that imports k8s.io/apimachinery
cat > types.go << 'EOF'
package test

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Example struct {
	metav1.TypeMeta
	Foo string
}
EOF

# Get latest version and tidy
go get k8s.io/apimachinery@latest
go mod tidy

# Get the resolved version
APIMACHINERY_VERSION=$(go list -m -f '{{.Version}}' k8s.io/apimachinery)
echo -e "${GREEN}Latest k8s.io/apimachinery version: $APIMACHINERY_VERSION${NC}"

# Copy files to embedded directory
echo "Updating embedded files..."
cp go.mod "$EMBEDDED_DIR/gomod.txt"
cp go.sum "$EMBEDDED_DIR/gosum.txt"

echo -e "${GREEN}✓ Successfully updated embedded go.mod and go.sum${NC}"
echo ""
echo "Files updated:"
echo "  - $EMBEDDED_DIR/gomod.txt"
echo "  - $EMBEDDED_DIR/gosum.txt"
echo ""
echo -e "${YELLOW}Note: Run tests to ensure compatibility:${NC}"
echo "  go test ./pkg/build/generation/crd -v"
echo ""
echo -e "${YELLOW}Commit the changes:${NC}"
echo "  git add pkg/build/generation/crd/embedded/"
echo "  git commit -m \"Update k8s.io/apimachinery to $APIMACHINERY_VERSION\""

