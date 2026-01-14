CREATE TABLE users (id INT PRIMARY KEY, name STRING);
INSERT INTO users (id, name) VALUES (1, 'Alice');
INSERT INTO users (id, name) VALUES (2, 'Bob');
SELECT * FROM users;
UPDATE users SET name = 'Bobby' WHERE id = 2;
SELECT * FROM users;
DELETE FROM users WHERE id = 1;
SELECT * FROM users;
exit
