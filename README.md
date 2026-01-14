# MiniBankDB

MiniBankDB is a custom relational database engine built from scratch in Go. It features a SQL-like query language, a B-Tree/Heap storage engine, and a web-based administration console.

## Features

- **SQL Support**: CREATE TABLE, INSERT, SELECT, UPDATE, DELETE, JOIN.
- **Constraints**: PRIMARY KEY and UNIQUE constraint enforcement.
- **Storage**: Page-based heap file storage with variable-length tuple support.
- **Indexing**: Hash Index support for O(1) equality lookups (created via `CREATE INDEX`).
- **Interfaces**: CLI REPL and Web Dashboard.

## Project Structure

- `db/`: Core database engine written in Go.
  - `cmd/minibank`: Entry point.
  - `internal/`: Internal modules (parser, planner, execution, storage).
- `web/ui`: Next.js frontend for the web console.
- `examples/`: SQL scripts for schema and seeding.
- `scripts/`: Helper scripts for running the project.

## Quickstart

### Prerequisites

- Go 1.25+
- Node.js & pnpm (for Web UI)

### 1. Run the Web Demo (Recommended)

This starts both the Go backend and Next.js frontend in a single command.

```bash
./scripts/run-web.sh
```

- **Frontend**: [http://localhost:3000](http://localhost:3000)
- **Backend**: [http://localhost:8080](http://localhost:8080)

**Note**: The web demo connects to the backend API at `NEXT_PUBLIC_API_URL` (default: `http://localhost:8080`). Data is persisted in the directory specified by the backend's `-data` flag (default: `./data`), including the catalog (`catalog.json`) and table data files.

### 2. Run the CLI REPL

For low-level SQL testing:

```bash
./scripts/run-repl.sh
```

### 3. Verify Persistence

Index definitions are persisted and in-memory indexes are automatically rebuilt on startup (no manual re-creation required). To verify:

1. Create a table and index in REPL.
2. Exit and restart REPL.
3. Check index usage with `SELECT * FROM table WHERE id = X`.
   (A helper script `tests/verify_persistence.sh` is provided).

## Architecture

1. **SQL Parser**: Recursive descent parser converting SQL to AST.
2. **Planner**: Converts AST to a tree of Execution Operators (Volcano Model).
3. **Execution Engine**: physical operators (SeqScan, IndexScan, Filter, Project, NestedLoopJoin).
4. **Storage Engine**: Manages data persistence using paging and heap files.

## Known Limitations (Notes)

1. **Indexing**: Indices are in-memory structures. While definitions are persisted to `catalog.json`, the index data structures are **automatically rebuilt from the heap file on server startup**. Large tables may penalize startup time.
2. **Concurrency**: The system uses basic table-level locking via Go mutexes. It is not designed for high-concurrency production workloads.
3. **Transactions**: There is no WAL (Write-Ahead Log) or ACID transaction support.
4. **Constraint Checking**: `PRIMARY KEY` and `UNIQUE` checks are optimized to use in-memory indices if available; otherwise, they fall back to a linear table scan at `INSERT` time.

## Test Cases

To verify the core "Systems" features, you can run the following SQL sequences in the REPL:

### 1. Constraint Enforcement

```sql
CREATE TABLE users (id INT PRIMARY KEY, name STRING UNIQUE);
INSERT INTO users (id, name) VALUES (1, 'Alice');
-- Fails with "duplicate primary key constraint violation"
INSERT INTO users (id, name) VALUES (1, 'Bob');
-- Fails with "unique constraint violation"
INSERT INTO users (id, name) VALUES (2, 'Alice');
```

### 2. Index Usage

```sql
CREATE INDEX idx_id ON users (id);
-- The planner will output: "[Planner] Using IndexScan on users.id"
SELECT * FROM users WHERE id = 1;
```

### 3. Join Support

```sql
CREATE TABLE accounts (acc_id INT, user_id INT, bal DECIMAL);
INSERT INTO accounts (acc_id, user_id, bal) VALUES (101, 1, 500.00);
SELECT * FROM users JOIN accounts ON users.id = accounts.user_id;
```

## License

MIT
