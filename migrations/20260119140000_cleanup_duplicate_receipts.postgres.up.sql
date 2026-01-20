-- 1. Eliminar recibos no autorizados si ya existe uno autorizado para la misma transacción
DELETE FROM electronic_receipts
WHERE id IN (
    SELECT r1.id
    FROM electronic_receipts r1
    JOIN electronic_receipts r2 ON r1.transaction_id = r2.transaction_id
    WHERE r1.sri_status != 'AUTORIZADO'
    AND r2.sri_status = 'AUTORIZADO'
    AND r1.id != r2.id
);

-- 2. Eliminar duplicados pendientes antiguos (mantener solo el más reciente por transacción si no hay autorizados)
DELETE FROM electronic_receipts
WHERE id IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER (PARTITION BY transaction_id ORDER BY created_at DESC) as rn
        FROM electronic_receipts
        WHERE transaction_id NOT IN (SELECT transaction_id FROM electronic_receipts WHERE sri_status = 'AUTORIZADO')
    ) t
    WHERE t.rn > 1
);
