-- Schema Definition for MiniBankDB

CREATE TABLE users (
    id INT PRIMARY KEY,
    name STRING,
    email STRING,
    created_at TIMESTAMP
);

CREATE TABLE accounts (
    id INT PRIMARY KEY,
    user_id INT,
    account_number STRING,
    balance DECIMAL,
    is_active BOOL
);

CREATE INDEX idx_users_name ON users (name);
CREATE INDEX idx_accounts_user_id ON accounts (user_id);
