#!/bin/bash
set -euo pipefail

cleanup() {
  rm -rf "$DATA_DIR"
}
trap cleanup EXIT

PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)
DATA_DIR="$PROJECT_ROOT/data-test"
mkdir -p "$DATA_DIR"
rm -f "$DATA_DIR/catalog.json"
rm -f "$DATA_DIR/users.data"

cd "$PROJECT_ROOT/db"

echo "Step 1: Create Table and Index..."
echo "
CREATE TABLE users (id INT PRIMARY KEY, name STRING);
CREATE INDEX idx_id ON users (id);
INSERT INTO users (id, name) VALUES (1, 'Alice');
INSERT INTO users (id, name) VALUES (2, 'Bob');
exit
" | go run cmd/minibank/main.go -data "$DATA_DIR"

echo "Step 2: Restart and Verify Index Usage..."
OUTPUT=$(echo "
SELECT * FROM users WHERE id = 1;
exit
" | go run cmd/minibank/main.go -data "$DATA_DIR" 2>&1)

if echo "$OUTPUT" | grep -q "Rebuilding indices"; then
    echo "PASS: Index rebuild triggered."
else
    echo "FAIL: Index rebuild not triggered."
    echo "Output was:"
    echo "$OUTPUT"
    exit 1
fi

echo "Persistence Verification Passed!"
rm -rf "$DATA_DIR"
