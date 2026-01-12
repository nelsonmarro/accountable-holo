CREATE TABLE emission_points (
  id SERIAL PRIMARY KEY,
  issuer_id INT NOT NULL,
  establishment_code VARCHAR(3) NOT NULL, -- Ej: 001
  emission_point_code VARCHAR(3) NOT NULL, -- Ej: 001
  receipt_type VARCHAR(2) NOT NULL, -- 01: Factura, 04: Nota de Crédito, etc.
  current_sequence INT NOT NULL DEFAULT 0, -- El último número usado
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  FOREIGN KEY (issuer_id) REFERENCES issuers (id) ON DELETE CASCADE,
  -- No permitir duplicados del mismo tipo de documento en el mismo punto para el mismo emisor
  UNIQUE (
    issuer_id,
    establishment_code,
    emission_point_code,
    receipt_type
  )
);
