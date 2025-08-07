ALTER TABLE transactions
DROP CONSTRAINT fk_transactions_created_by,
DROP CONSTRAINT fk_transactions_updated_by;

ALTER TABLE transactions
DROP COLUMN created_by_id,
DROP COLUMN updated_by_id;
