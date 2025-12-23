-- Revert 'Ajuste por Reconciliación'
UPDATE categories 
SET name = 'Ajuste por ReconciliaciÃ³n' 
WHERE name = 'Ajuste por Reconciliación';

-- Revert 'Anular Transacción'
UPDATE categories 
SET name = 'Anular TransacciÃ³n I' 
WHERE name = 'Anular Transacción Ingreso';

UPDATE categories 
SET name = 'Anular TransacciÃ³n E' 
WHERE name = 'Anular Transacción Egreso';
