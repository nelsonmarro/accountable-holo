DROP TABLE transaction_items;

ALTER TABLE transactions 
    DROP CONSTRAINT fk_transactions_tax_payers,
    DROP COLUMN subtotal_15,
    DROP COLUMN subtotal_0,
    DROP COLUMN tax_amount,
    DROP COLUMN tax_payer_id;
