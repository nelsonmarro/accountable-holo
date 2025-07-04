ALTER TABLE transactions
DROP COLUMN IF EXISTS is_voided;

ALTER TABLE transactions
DROP COLUMN IF EXISTS voided_by_transaction_id;

ALTER TABLE transactions
DROP COLUMN IF EXISTS voids_transaction_id;
