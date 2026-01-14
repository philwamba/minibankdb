# MiniBankDB Web Demo

MiniBankDB now includes a "Trivial Web App" dashboard to demonstrate application-level CRUD operations powered by the custom Go RDBMS engine.

## Features

- **Users Management**: Create, Read, Update, Delete (CRUD) users.
- **Wallet Management**: Manage user wallets and balances.
- **Transactions**: Record deposits, withdrawals, and transfers.
- **Reporting**: View a joined report of Users and Wallets.
- **SQL Console**: Execute raw SQL queries directly against the engine.

## Architecture

- **Frontend**: Next.js (App Router) + Tailwind CSS + shadcn/ui.
- **Backend**: Go HTTP Server (`db/cmd/minibank`) interacting directly with `planner`/`execution` layers.
- **Database**: MiniBankDB Custom Engine (HeapFile storage, Hash Indexing).

## Usage

1. **Start the Demo**:
   Run the all-in-one script from the project root:

   ```bash
   ./scripts/run-web.sh
   ```

2. **Access the Application**:
   - **Frontend UI**: [http://localhost:3000](http://localhost:3000)
   - **Backend API**: [http://localhost:8080](http://localhost:8080)

3. **Verify Functionality**:
   - Navigate to **Users** and create a new user.
   - Go to **Wallets** and assign a wallet to that user ID.
   - Go to **Reports** to see the joined data.
   - Restart the server (`Ctrl+C` then run script again) and verify data persists.

## Technical Details

- **Constraint Checking**: Uses `HashIndex` for duplicate detection (O(1)) instead of scan (O(N)).
- **Persistence**: Table schema and index definitions are persisted in `catalog.json`. Indexes are automatically rebuilt on startup.
