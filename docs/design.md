# MiniBankDB Design Document

## Overview

MiniBankDB is a specialized relational database designed to demonstrate core database systems concepts.

## Storage Format

### Heap Files

Tables are stored as "Heap Files" (`.data`). Each file consists of 4KB pages.

### Page Layout (Slotted Page)

We use a **Slotted Page** structure to handle variable-length records efficiently within a fixed-size page.

- **Header**: Contains page metadata (page ID, slot count, free space pointer).
- **Slots**: Array of pointers (offset, length) growing from the header downwards.
- **Tuples**: Data records growing from the end of the page upwards.

### Tuple Format

Tuples are serialized binary data:

- `[Cell Count] [Cell Type][Cell Length][Value] ...`

## Query Processing

### Parser

Custom lexer and recursive descent parser support a subset of SQL-92. Defaulting to standard precedence rules.

### Execution

Uses the **Volcano Iterator Model**. Each operator implements `Open()`, `Next()`, `Close()`.

- `Next()` passes control down the tree, pulling tuples one by one.
- This allows for pipelined execution and low memory overhead.

## Indexing

Currently supports **Hash Indexing** for O(1) equality lookups. The index is built in-memory on startup (or on demand) mapping `Key -> []RID`.

## Concurrency

Basic thread-safety using Go's `sync.Mutex` at the file/page level.

## Web Interface

Exposes a simple REST API (`POST /api/query`) wrapped by a modern Next.js dashboard.
