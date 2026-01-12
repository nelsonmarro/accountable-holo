-- 1. Añadir columnas de desglose a transactions
ALTER TABLE transactions 
    ADD COLUMN subtotal_15 NUMERIC(15, 2) DEFAULT 0 NOT NULL,
    ADD COLUMN subtotal_0 NUMERIC(15, 2) DEFAULT 0 NOT NULL,
    ADD COLUMN tax_amount NUMERIC(15, 2) DEFAULT 0 NOT NULL,
    ADD COLUMN tax_payer_id INT; -- Opcional, puede ser NULL (Consumidor Final)

-- Agregar la clave foránea para el cliente
ALTER TABLE transactions
    ADD CONSTRAINT fk_transactions_tax_payers
    FOREIGN KEY (tax_payer_id) REFERENCES tax_payers (id);

-- 2. Crear tabla de detalles (Items)
CREATE TABLE transaction_items (
    id SERIAL PRIMARY KEY,
    transaction_id INT NOT NULL,
    description TEXT NOT NULL,
    quantity NUMERIC(15, 6) NOT NULL DEFAULT 1, -- Hasta 6 decimales según SRI
    unit_price NUMERIC(15, 6) NOT NULL,         -- Precio unitario sin impuestos
    tax_rate INT NOT NULL DEFAULT 0,            -- Código de impuesto (0, 2, 4)
    subtotal NUMERIC(15, 2) NOT NULL,           -- quantity * unit_price
    
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    FOREIGN KEY (transaction_id) REFERENCES transactions (id) ON DELETE CASCADE
);
