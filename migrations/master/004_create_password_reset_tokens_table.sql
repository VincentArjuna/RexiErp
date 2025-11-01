-- Migration: Create password_reset_tokens table
-- Created: 2025-11-02
-- Description: Table for storing password reset tokens with security features

-- Enable UUID extension if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create password_reset_tokens table
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    token VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE NULL,
    ip_address VARCHAR(45) NULL,
    user_agent TEXT NULL,
    is_active BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Create indexes for performance and constraints
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_tenant_id ON password_reset_tokens(tenant_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token_hash ON password_reset_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_used_at ON password_reset_tokens(used_at);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_is_active ON password_reset_tokens(is_active);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_deleted_at ON password_reset_tokens(deleted_at);

-- Add unique constraints for security
ALTER TABLE password_reset_tokens ADD CONSTRAINT password_reset_tokens_token_unique UNIQUE (token);
ALTER TABLE password_reset_tokens ADD CONSTRAINT password_reset_tokens_token_hash_unique UNIQUE (token_hash);

-- Add foreign key constraints
ALTER TABLE password_reset_tokens ADD CONSTRAINT password_reset_tokens_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE password_reset_tokens ADD CONSTRAINT password_reset_tokens_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

-- Add check constraints
ALTER TABLE password_reset_tokens ADD CONSTRAINT password_reset_tokens_expires_at_check
    CHECK (expires_at > created_at);

ALTER TABLE password_reset_tokens ADD CONSTRAINT password_reset_tokens_token_length_check
    CHECK (length(token) >= 10);

ALTER TABLE password_reset_tokens ADD CONSTRAINT password_reset_tokens_token_hash_length_check
    CHECK (length(token_hash) >= 32);

-- Add trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_password_reset_tokens_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER password_reset_tokens_updated_at_trigger
    BEFORE UPDATE ON password_reset_tokens
    FOR EACH ROW
    EXECUTE FUNCTION update_password_reset_tokens_updated_at();

-- Add comments for documentation
COMMENT ON TABLE password_reset_tokens IS 'Stores password reset tokens for users with security features';
COMMENT ON COLUMN password_reset_tokens.id IS 'Unique identifier for the password reset token';
COMMENT ON COLUMN password_reset_tokens.user_id IS 'Foreign key to the user requesting password reset';
COMMENT ON COLUMN password_reset_tokens.tenant_id IS 'Foreign key to the tenant for multi-tenant isolation';
COMMENT ON COLUMN password_reset_tokens.token IS 'The raw reset token (should be kept secure)';
COMMENT ON COLUMN password_reset_tokens.token_hash IS 'Hashed version of the token for secure lookup';
COMMENT ON COLUMN password_reset_tokens.email IS 'Email address where reset token was sent';
COMMENT ON COLUMN password_reset_tokens.expires_at IS 'When the token expires and becomes invalid';
COMMENT ON COLUMN password_reset_tokens.used_at IS 'When the token was used to reset password';
COMMENT ON COLUMN password_reset_tokens.ip_address IS 'IP address of the requester for security auditing';
COMMENT ON COLUMN password_reset_tokens.user_agent IS 'User agent of the requester for security auditing';
COMMENT ON COLUMN password_reset_tokens.is_active IS 'Whether the token is currently active';
COMMENT ON COLUMN password_reset_tokens.created_at IS 'When the token was created';
COMMENT ON COLUMN password_reset_tokens.updated_at IS 'When the token was last updated';
COMMENT ON COLUMN password_reset_tokens.deleted_at IS 'When the token was soft deleted';

-- Create view for active tokens only (useful for reporting)
CREATE OR REPLACE VIEW active_password_reset_tokens AS
SELECT
    id,
    user_id,
    tenant_id,
    email,
    expires_at,
    ip_address,
    user_agent,
    created_at,
    updated_at
FROM password_reset_tokens
WHERE is_active = true
  AND deleted_at IS NULL
  AND used_at IS NULL
  AND expires_at > CURRENT_TIMESTAMP;

COMMENT ON VIEW active_password_reset_tokens IS 'View of currently active password reset tokens';