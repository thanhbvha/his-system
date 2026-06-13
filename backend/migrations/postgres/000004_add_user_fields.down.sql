ALTER TABLE users 
DROP COLUMN email_encrypted,
DROP COLUMN email_hmac,
DROP COLUMN mfa_enabled;
