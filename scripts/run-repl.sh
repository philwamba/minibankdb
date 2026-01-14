#!/bin/bash
PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$PROJECT_ROOT/db/cmd/minibank"
rm -f "$PROJECT_ROOT/data/"*.data
rm -f "$PROJECT_ROOT/data/"catalog.json
go run main.go -mode repl -data "$PROJECT_ROOT/data"
