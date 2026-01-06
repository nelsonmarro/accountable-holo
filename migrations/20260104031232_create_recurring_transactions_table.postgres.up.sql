CREATE TABLE recurring_transactions (
  id SERIAL PRIMARY KEY,
  description TEXT NOT NULL,
  amount NUMERIC(15, 2) NOT NULL,
  account_id INT NOT NULL,
  category_id INT NOT NULL,

  -- Configuración de Recurrencia
  interval VARCHAR(20) NOT NULL, -- 'MONTHLY', 'WEEKLY', etc.
  start_date DATE NOT NULL,
  next_run_date DATE NOT NULL,
  is_active BOOLEAN DEFAULT TRUE,

  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,

  -- Claves foráneas (Integridad)
  FOREIGN KEY (account_id) REFERENCES accounts (id) ON DELETE RESTRICT,
  FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE RESTRICT
);
