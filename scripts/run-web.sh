#!/bin/bash
set -euo pipefail

# Define cleanup function to kill background processes
cleanup() {
    echo "Stopping MiniBankDB Web Demo..."
    kill $(jobs -p) 2>/dev/null || true
}
trap cleanup EXIT

PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)
WEB_UI_DIR="$PROJECT_ROOT/web/ui"
BACKEND_PORT="${BACKEND_PORT:-8080}"
FRONTEND_PORT="${FRONTEND_PORT:-3000}"

echo "Starting MiniBankDB Backend on port $BACKEND_PORT..."
cd "$PROJECT_ROOT/db/cmd/minibank"
go run main.go -mode server -port :$BACKEND_PORT -data "$PROJECT_ROOT/data" &
BACKEND_PID=$!

echo "Waiting for backend to initialize..."
sleep 2

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
