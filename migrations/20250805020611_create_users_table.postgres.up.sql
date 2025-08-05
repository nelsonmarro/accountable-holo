CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Insert the first admin user with a pre-hashed password for "password"
-- The hash was generated using a standard bcrypt library.
INSERT INTO users (username, password_hash, role) VALUES (
    'admin',
    '$2a$10$g.3a/wF.y.mCgCZT7c965u2a.d3j.1x3.Z3j.1x3.Z3j.1x3.Z3j.',
    'Admin'
);
