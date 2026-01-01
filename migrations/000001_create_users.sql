-- +migrate UP
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_users_email ON users(email);

-- +migrate DOWN
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
