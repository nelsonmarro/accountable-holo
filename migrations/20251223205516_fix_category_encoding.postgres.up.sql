-- Fix 'Ajuste por Reconciliación'
UPDATE categories 
SET name = 'Ajuste por Reconciliación' 
WHERE name LIKE 'Ajuste por Reconciliaci%';

-- Fix 'Anular Transacción'
UPDATE categories 
SET name = 'Anular Transacción Ingreso' 
WHERE name LIKE 'Anular Transacci%n I%';

UPDATE categories 
SET name = 'Anular Transacción Egreso' 
WHERE name LIKE 'Anular Transacci%n E%';
