ALTER TABLE transactions
ADD COLUMN type VARCHAR(100) DEFAULT '' NOT NULL;

ALTER TABLE categories
DROP COLUMN type;
