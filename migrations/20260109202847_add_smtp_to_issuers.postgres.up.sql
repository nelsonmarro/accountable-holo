ALTER TABLE issuers
    ADD COLUMN smtp_server VARCHAR(255),
    ADD COLUMN smtp_port INT,
    ADD COLUMN smtp_user VARCHAR(255),
    ADD COLUMN smtp_password VARCHAR(255),
    ADD COLUMN smtp_ssl BOOLEAN DEFAULT TRUE;
