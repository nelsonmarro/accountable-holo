ALTER TABLE transactions
ADD COLUMN transaction_number VARCHAR(20) NOT NULL;

ALTER TABLE transactions
ADD CONSTRAINT uq_transaction_number UNIQUE (transaction_number);
