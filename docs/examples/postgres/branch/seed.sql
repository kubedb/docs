CREATE TABLE IF NOT EXISTS branch_demo (
    id integer PRIMARY KEY,
    email text NOT NULL
);
INSERT INTO branch_demo (id, email)
VALUES (1, 'alice@example.com'), (2, 'bob@example.com')
ON CONFLICT (id) DO NOTHING;
