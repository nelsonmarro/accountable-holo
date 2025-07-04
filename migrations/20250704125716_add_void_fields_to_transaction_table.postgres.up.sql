ALTER TABLE transactions
ADD COLUMN is_voided BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE transactions
ADD COLUMN voided_by_transaction_id INTEGER REFERENCES transactions(id) NULL;

ALTER TABLE transactions
ADD COLUMN voids_transaction_id INTEGER REFERENCES transactions(id) NULL;
