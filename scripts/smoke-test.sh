#!/bin/bash
# Binarius Smoke Test Suite
# Runs end-to-end tests on a built binary to verify core functionality
#
# Usage: ./scripts/smoke-test.sh [binary-path]
# Example: ./scripts/smoke-test.sh ./binarius-linux-amd64

set -euo pipefail

BINARY="${1:-./binarius}"

# Verify binary exists
if [[ ! -x "$BINARY" ]]; then
    echo "Error: Binary not found or not executable: $BINARY"
    exit 1
fi

# Setup isolated test environment
TEST_HOME=$(mktemp -d)
export HOME="$TEST_HOME"
export PATH="$HOME/.local/bin:$PATH"
mkdir -p "$HOME/.local/bin"

# Cleanup on exit
cleanup() {
    rm -rf "$TEST_HOME"
}
trap cleanup EXIT

echo "============================================"
echo "Binarius Smoke Test Suite"
echo "============================================"
echo "Binary: $BINARY"
echo "Test HOME: $TEST_HOME"
echo ""

# Test 1: Basic execution
echo "--- Test: Basic execution ---"
$BINARY --version
$BINARY --help >/dev/null
echo "PASS: Binary executes correctly"
echo ""

# Test 2: Init command
echo "--- Test: Init command ---"
$BINARY init
test -d "$HOME/.binarius" || { echo "FAIL: .binarius not created"; exit 1; }
test -d "$HOME/.binarius/tools" || { echo "FAIL: tools dir not created"; exit 1; }
test -d "$HOME/.binarius/cache" || { echo "FAIL: cache dir not created"; exit 1; }
test -f "$HOME/.binarius/config.yaml" || { echo "FAIL: config.yaml not created"; exit 1; }
test -f "$HOME/.binarius/installation.json" || { echo "FAIL: installation.json not created"; exit 1; }
echo "PASS: Init creates correct directory structure"
echo ""

# Test 3: Install terraform
echo "--- Test: Install terraform ---"
$BINARY install terraform@v1.5.7
echo "PASS: Terraform installed"
echo ""

# Test 4: List command
echo "--- Test: List command ---"
$BINARY list terraform
echo "PASS: List works correctly"
echo ""

# Test 5: Use command (symlink creation)
echo "--- Test: Use command ---"
$BINARY use terraform@v1.5.7
test -L "$HOME/.local/bin/terraform" || { echo "FAIL: Symlink not created"; exit 1; }
echo "PASS: Symlink created"
echo ""

# Test 6: Info command (requires active version)
echo "--- Test: Info command ---"
$BINARY info terraform
echo "PASS: Info works correctly"
echo ""

# Test 7: Verify tool executes via symlink
echo "--- Test: Tool execution via symlink ---"
terraform version
echo "PASS: Terraform executes via symlink"
echo ""

# Test 8: Uninstall
echo "--- Test: Uninstall command ---"
$BINARY uninstall terraform@v1.5.7 --force
echo "PASS: Uninstall completed"
echo ""

echo "============================================"
echo "All smoke tests passed!"
echo "============================================"
