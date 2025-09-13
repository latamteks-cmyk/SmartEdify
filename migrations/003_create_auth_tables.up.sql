-- Create refresh_tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jti VARCHAR(255) UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    unit_id UUID REFERENCES units(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    revoked_reason VARCHAR(255),
    device_info JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create webauthn_credentials table for FIDO2 support
CREATE TABLE IF NOT EXISTS webauthn_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id TEXT UNIQUE NOT NULL,
    public_key TEXT NOT NULL,
    attestation_type VARCHAR(50),
    transport VARCHAR(255),
    flags JSONB DEFAULT '{}',
    counter BIGINT DEFAULT 0,
    device_name VARCHAR(255),
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create login_attempts table for rate limiting and security
CREATE TABLE IF NOT EXISTS login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identifier VARCHAR(255) NOT NULL, -- email, phone, or IP
    identifier_type VARCHAR(50) NOT NULL CHECK (identifier_type IN ('email', 'phone', 'ip')),
    attempt_type VARCHAR(50) NOT NULL CHECK (attempt_type IN ('password', 'otp', 'webauthn')),
    success BOOLEAN NOT NULL,
    ip_address INET,
    user_agent TEXT,
    error_reason VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create otp_codes table for WhatsApp/SMS verification
CREATE TABLE IF NOT EXISTS otp_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone VARCHAR(20) NOT NULL,
    code VARCHAR(10) NOT NULL,
    code_hash VARCHAR(255) NOT NULL,
    purpose VARCHAR(50) NOT NULL CHECK (purpose IN ('login', 'registration', 'password_reset', 'mfa')),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP WITH TIME ZONE,
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    ip_address INET,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create audit_logs table for immutable audit trail
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    event_category VARCHAR(50) NOT NULL CHECK (event_category IN ('auth', 'user', 'admin', 'compliance', 'security')),
    user_id UUID REFERENCES users(id),
    tenant_id UUID REFERENCES tenants(id),
    unit_id UUID REFERENCES units(id),
    resource_type VARCHAR(50),
    resource_id VARCHAR(255),
    action VARCHAR(100) NOT NULL,
    details JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(255),
    correlation_id VARCHAR(255),
    hash_chain VARCHAR(255),
    previous_hash VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_refresh_tokens_jti ON refresh_tokens(jti);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_revoked ON refresh_tokens(revoked);

CREATE INDEX idx_webauthn_credentials_user_id ON webauthn_credentials(user_id);
CREATE INDEX idx_webauthn_credentials_credential_id ON webauthn_credentials(credential_id);

CREATE INDEX idx_login_attempts_identifier ON login_attempts(identifier);
CREATE INDEX idx_login_attempts_identifier_type ON login_attempts(identifier_type);
CREATE INDEX idx_login_attempts_created_at ON login_attempts(created_at);
CREATE INDEX idx_login_attempts_ip_address ON login_attempts(ip_address);

CREATE INDEX idx_otp_codes_phone ON otp_codes(phone);
CREATE INDEX idx_otp_codes_expires_at ON otp_codes(expires_at);
CREATE INDEX idx_otp_codes_used ON otp_codes(used);

CREATE INDEX idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_logs_event_category ON audit_logs(event_category);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_correlation_id ON audit_logs(correlation_id);

-- Create updated_at triggers
CREATE TRIGGER update_refresh_tokens_updated_at BEFORE UPDATE ON refresh_tokens
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_webauthn_credentials_updated_at BEFORE UPDATE ON webauthn_credentials
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function to maintain hash chain for audit logs
CREATE OR REPLACE FUNCTION maintain_audit_hash_chain()
RETURNS TRIGGER AS $$
DECLARE
    last_hash VARCHAR(255);
BEGIN
    -- Get the last hash from the most recent audit log
    SELECT hash_chain INTO last_hash
    FROM audit_logs
    ORDER BY created_at DESC
    LIMIT 1;
    
    -- Set the previous hash
    NEW.previous_hash = COALESCE(last_hash, '');
    
    -- Calculate the new hash (simplified - in production use proper cryptographic hash)
    NEW.hash_chain = encode(sha256((NEW.id::text || NEW.event_type || NEW.created_at::text || NEW.previous_hash)::bytea), 'hex');
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for audit log hash chain
CREATE TRIGGER maintain_audit_hash_chain_trigger
    BEFORE INSERT ON audit_logs
    FOR EACH ROW EXECUTE FUNCTION maintain_audit_hash_chain();