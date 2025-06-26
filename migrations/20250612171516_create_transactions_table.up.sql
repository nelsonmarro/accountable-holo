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
