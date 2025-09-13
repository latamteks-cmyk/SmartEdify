-- Drop triggers and functions
DROP TRIGGER IF EXISTS maintain_audit_hash_chain_trigger ON audit_logs;
DROP FUNCTION IF EXISTS maintain_audit_hash_chain();

DROP TRIGGER IF EXISTS update_webauthn_credentials_updated_at ON webauthn_credentials;
DROP TRIGGER IF EXISTS update_refresh_tokens_updated_at ON refresh_tokens;

-- Drop tables
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS otp_codes;
DROP TABLE IF EXISTS login_attempts;
DROP TABLE IF EXISTS webauthn_credentials;
DROP TABLE IF EXISTS refresh_tokens;