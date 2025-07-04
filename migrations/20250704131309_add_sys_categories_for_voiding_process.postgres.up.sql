INSERT INTO categories (name, type, created_at, updated_at)
VALUES
('Anular Transacción E', 'Egreso', NOW(), NOW()),
('Anular Transacción I', 'Ingreso', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;
