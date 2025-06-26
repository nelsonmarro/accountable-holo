ALTER TABLE transactions ADD CONSTRAINT transactions_account_fk FOREIGN KEY (account_id) REFERENCES accounts (id) ON DELETE CASCADE ON UPDATE CASCADE;
