INSERT INTO tax_payers (identification, identification_type, name, email, created_at, updated_at)
VALUES ('9999999999999', '07', 'CONSUMIDOR FINAL', 'consumidorfinal@verith.com', NOW(), NOW())
ON CONFLICT (identification) DO NOTHING;
