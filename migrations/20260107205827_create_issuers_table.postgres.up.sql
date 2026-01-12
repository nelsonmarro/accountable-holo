CREATE TABLE issuers (
  id SERIAL PRIMARY KEY,
  ruc VARCHAR(13) NOT NULL UNIQUE,
  business_name VARCHAR(300) NOT NULL,
  trade_name VARCHAR(300),
  main_address VARCHAR(300) NOT NULL,
  establishment_address VARCHAR(300) NOT NULL,
  establishment_code VARCHAR(3) NOT NULL,
  emission_point_code VARCHAR(3) NOT NULL,
  contribution_class VARCHAR(50),
  withholding_agent VARCHAR(50),
  rimpe_type VARCHAR(50),
  environment INT NOT NULL DEFAULT 1,
  signature_path TEXT NOT NULL,
  logo_path TEXT,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);
