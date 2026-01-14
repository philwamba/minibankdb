CREATE TABLE users (
    id INT PRIMARY KEY,
    username STRING UNIQUE,
    email STRING
);

INSERT INTO users (id, username, email) VALUES (1, 'alice', 'alice@example.com');
-- Should succeed
INSERT INTO users (id, username, email) VALUES (2, 'bob', 'bob@example.com');

-- Should fail (Duplicate PK)
INSERT INTO users (id, username, email) VALUES (1, 'charlie', 'charlie@example.com');

-- Should fail (Duplicate Unique)
INSERT INTO users (id, username, email) VALUES (3, 'alice', 'alice2@example.com');

SELECT * FROM users;
