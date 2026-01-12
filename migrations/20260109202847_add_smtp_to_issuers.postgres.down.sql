ALTER TABLE issuers
    DROP COLUMN smtp_server,
    DROP COLUMN smtp_port,
    DROP COLUMN smtp_user,
    DROP COLUMN smtp_password,
    DROP COLUMN smtp_ssl;
