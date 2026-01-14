CREATE TABLE indexed_users (
    id INT PRIMARY KEY,
    name STRING
);

INSERT INTO indexed_users (id, name) VALUES (1, 'alice');
INSERT INTO indexed_users (id, name) VALUES (2, 'bob');
INSERT INTO indexed_users (id, name) VALUES (3, 'charlie');

-- Create Index
CREATE INDEX idx_name ON indexed_users (name);

-- Select using index (Should print "[Planner] Using IndexScan...")
SELECT * FROM indexed_users WHERE name = 'bob';

-- Select NOT using index (SeqScan)
SELECT * FROM indexed_users WHERE id = 1;

-- Test duplicate prevention (PK)
INSERT INTO indexed_users (id, name) VALUES (4, 'david');
-- This next insert MUST FAIL with a constraint violation error
INSERT INTO indexed_users (id, name) VALUES (4, 'duplicate');
