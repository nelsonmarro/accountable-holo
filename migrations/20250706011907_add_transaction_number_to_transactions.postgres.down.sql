ALTER TABLE transactions
DROP CONSTRAINT uq_transaction_number;

ALTER TABLE transactions
DROP COLUMN transaction_number;
