CREATE TABLE tax_payers (
  id SERIAL PRIMARY KEY,
  identification VARCHAR(13) NOT NULL UNIQUE, -- RUC, Cédula o Pasaporte
  identification_type VARCHAR(2) NOT NULL, -- 04: RUC, 05: Cédula, 06: Pasaporte, 07: Consumidor Final, 08: Exterior
  name VARCHAR(300) NOT NULL, -- Razón Social o Nombre Completo
  email VARCHAR(300) NOT NULL, -- Vital para el envío electrónico
  address VARCHAR(300),
  phone VARCHAR(20),
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);
