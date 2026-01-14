#!/bin/bash
set -euo pipefail

# Define cleanup function to kill background processes
cleanup() {
    echo "Stopping MiniBankDB Web Demo..."
    # Read pids into array to avoid word splitting issues
    local pids
    if jobs -p >/dev/null 2>&1; then
        readarray -t pids < <(jobs -p)
        if [ ${#pids[@]} -gt 0 ]; then
            kill "${pids[@]}" 2>/dev/null || true
        fi
    fi
}
trap cleanup EXIT

PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)
WEB_UI_DIR="$PROJECT_ROOT/web/ui"
BACKEND_PORT="${BACKEND_PORT:-8080}"
FRONTEND_PORT="${FRONTEND_PORT:-3000}"

echo "Starting MiniBankDB Backend on port $BACKEND_PORT..."
cd "$PROJECT_ROOT/db/cmd/minibank"
go run main.go -mode server -port :$BACKEND_PORT -data "$PROJECT_ROOT/data" &
# We don't really need BACKEND_PID if we use jobs -p, but keeping it is harmless.

echo "Waiting for backend to initialize..."
for _ in {1..50}; do
  if curl -fsS "http://localhost:$BACKEND_PORT/api/wallets" >/dev/null 2>&1; then
    break
  fi
  if ! kill -0 $! 2>/dev/null; then
    echo "Backend failed to start"
    exit 1
  fi
  sleep 0.1
done

echo "Preparing Web UI..."
cd "$WEB_UI_DIR"
if [ ! -d "node_modules" ]; then
    echo "Installing frontend dependencies..."
    pnpm install
fi

echo "Starting Web UI on port $FRONTEND_PORT..."
echo "-----------------------------------------------------"
echo "MiniBankDB Web Demo is running!"
echo "Backend API: http://localhost:$BACKEND_PORT"
echo "Frontend UI: http://localhost:$FRONTEND_PORT"
echo "-----------------------------------------------------"

# Pass port to Next.js if possible, or assume default/env var
PORT=$FRONTEND_PORT pnpm dev
