#!/bin/bash
PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$PROJECT_ROOT/db/cmd/minibank"
go run main.go -mode server -port :8080 -data "$PROJECT_ROOT/data"
