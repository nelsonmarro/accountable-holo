CREATE TABLE categories (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) DEFAULT '' NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE accounts (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) DEFAULT '' NOT NULL,
  type VARCHAR(100) DEFAULT '' NOT NULL,
  initial_balance REAL NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);


CREATE TABLE transactions (
  id SERIAL PRIMARY KEY,
  amount REAL NOT NULL,
  type VARCHAR(100) DEFAULT '' NOT NULL,
  description VARCHAR(300) DEFAULT '',
  account_id INTEGER,
  category_id INTEGER,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

ALTER TABLE transactions ADD CONSTRAINT transactions_category_fk FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE transactions ADD CONSTRAINT transactions_account_fk FOREIGN KEY (account_id) REFERENCES accounts (id) ON DELETE CASCADE ON UPDATE CASCADE;

CREATE UNIQUE INDEX accounts_name_idx ON accounts (name);
CREATE UNIQUE INDEX categories_name_idx ON categories (name);

ALTER TABLE accounts ADD COLUMN number VARCHAR(100);

ALTER TABLE categories
ADD COLUMN type VARCHAR(50) NOT NULL DEFAULT 'outcome';

ALTER TABLE transactions
DROP COLUMN type;

ALTER TABLE transactions
ADD COLUMN is_voided BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE transactions
ADD COLUMN voided_by_transaction_id INTEGER REFERENCES transactions(id) NULL;

ALTER TABLE transactions
ADD COLUMN voids_transaction_id INTEGER REFERENCES transactions(id) NULL;

INSERT INTO categories (name, type, created_at, updated_at)
VALUES
('Anular Transacción E', 'Egreso', NOW(), NOW()),
('Anular Transacción I', 'Ingreso', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

ALTER TABLE transactions
ADD COLUMN transaction_date DATE NOT NULL DEFAULT CURRENT_DATE;

ALTER TABLE transactions
ALTER COLUMN amount TYPE NUMERIC(10, 2);

ALTER TABLE transactions
ADD COLUMN transaction_number VARCHAR(20) NOT NULL;

ALTER TABLE transactions
ADD CONSTRAINT uq_transaction_number UNIQUE (transaction_number);
ALTER TABLE transactions ADD COLUMN attachment_path TEXT;
INSERT INTO categories (name, type, created_at, updated_at) VALUES ('Ajuste por Reconciliación', 'Ajuste', NOW(), NOW());

-- Drop the old unique index on the name column
DROP INDEX categories_name_idx;

-- Create a new composite unique index on name and type
CREATE UNIQUE INDEX categories_name_type_idx ON categories (name, type);

INSERT INTO categories (name, type, created_at, updated_at) VALUES ('Ajuste por Reconciliación', 'Ingreso', NOW(), NOW());
INSERT INTO categories (name, type, created_at, updated_at) VALUES ('Ajuste por Reconciliación', 'Egreso', NOW(), NOW());

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

UPDATE users SET password_hash = '$2a$10$REAsEaVqya7iPJH/TyIV9.Klt8xNExOIKnyR62rsZlZoir0Zuapqm' WHERE username = 'admin';

ALTER TABLE users ADD COLUMN first_name VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN last_name VARCHAR(255) NOT NULL DEFAULT '';

UPDATE users SET first_name = 'Nelson', last_name = 'Marro' WHERE username = 'admin';

ALTER TABLE transactions
ADD COLUMN created_by_id INT,
ADD COLUMN updated_by_id INT;

ALTER TABLE transactions
ADD CONSTRAINT fk_transactions_created_by FOREIGN KEY (created_by_id) REFERENCES users(id),
ADD CONSTRAINT fk_transactions_updated_by FOREIGN KEY (updated_by_id) REFERENCES users(id);

