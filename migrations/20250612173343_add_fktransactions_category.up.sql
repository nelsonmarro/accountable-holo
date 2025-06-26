ALTER TABLE transactions ADD CONSTRAINT transactions_category_fk FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE CASCADE ON UPDATE CASCADE;
