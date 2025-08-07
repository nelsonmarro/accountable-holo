ALTER TABLE transactions
ADD COLUMN created_by_id INT,
ADD COLUMN updated_by_id INT;

ALTER TABLE transactions
ADD CONSTRAINT fk_transactions_created_by FOREIGN KEY (created_by_id) REFERENCES users(id),
ADD CONSTRAINT fk_transactions_updated_by FOREIGN KEY (updated_by_id) REFERENCES users(id);
