-- Seed Data

INSERT INTO users (id, name, email, created_at) VALUES (1, 'Alice Smith', 'alice@example.com', '2023-01-01T10:00:00Z');
INSERT INTO users (id, name, email, created_at) VALUES (2, 'Bob Jones', 'bob@example.com', '2023-01-02T11:30:00Z');
INSERT INTO users (id, name, email, created_at) VALUES (3, 'Charlie Brown', 'charlie@example.com', '2023-01-03T09:15:00Z');

INSERT INTO accounts (id, user_id, account_number, balance, is_active) VALUES (101, 1, 'ACC-1001', 5000.50, true);
INSERT INTO accounts (id, user_id, account_number, balance, is_active) VALUES (102, 2, 'ACC-1002', 150.00, true);
INSERT INTO accounts (id, user_id, account_number, balance, is_active) VALUES (103, 1, 'ACC-1003', 0.00, false);
