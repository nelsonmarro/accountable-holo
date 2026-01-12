CREATE TABLE electronic_receipts (
  id SERIAL PRIMARY KEY,
  transaction_id INT NOT NULL,
  issuer_id INT NOT NULL,
  tax_payer_id INT NOT NULL,
  -- Identificación única del SRI (49 dígitos)
  access_key VARCHAR(49) NOT NULL UNIQUE,
  -- Detalles del Comprobante
  receipt_type VARCHAR(2) NOT NULL, -- 01: Factura, 04: Nota de Crédito
  xml_content TEXT, -- Aquí guardaremos el XML firmado completo
  authorization_date TIMESTAMP, -- Fecha/Hora de autorización del SRI
  -- Estados: PENDIENTE, RECIBIDA, AUTORIZADO, RECHAZADO, DEVUELTA
  sri_status VARCHAR(20) NOT NULL DEFAULT 'PENDIENTE',
  sri_message TEXT, -- Mensaje de error o éxito del SRI
  -- Gestión Local
  ride_path TEXT, -- Ubicación del PDF generado (RIDE)
  environment INT NOT NULL, -- 1: Pruebas, 2: Producción
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  -- Relaciones e Integridad
  FOREIGN KEY (transaction_id) REFERENCES transactions (id) ON DELETE CASCADE,
  FOREIGN KEY (issuer_id) REFERENCES issuers (id) ON DELETE RESTRICT,
  FOREIGN KEY (tax_payer_id) REFERENCES tax_payers (id) ON DELETE RESTRICT
);
