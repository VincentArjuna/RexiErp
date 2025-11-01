-- Migration: Create user_sessions table
-- Created: Authentication Service
-- Description: Creates the user_sessions table for token management and session tracking

-- Create user_sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    refresh_token_hash VARCHAR(255) NOT NULL,
    device_info JSON,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    last_activity TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance and security
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_tenant_id ON user_sessions(tenant_id);
CREATE UNIQUE INDEX idx_user_sessions_session_id ON user_sessions(session_id);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_user_sessions_last_activity ON user_sessions(last_activity);
CREATE INDEX idx_user_sessions_is_active ON user_sessions(is_active);
CREATE INDEX idx_user_sessions_created_at ON user_sessions(created_at);

-- Create trigger for updated_at timestamp
CREATE TRIGGER update_user_sessions_updated_at
    BEFORE UPDATE ON user_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add foreign key constraints
ALTER TABLE user_sessions
    ADD CONSTRAINT fk_user_sessions_user_id
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add table comments for documentation
COMMENT ON TABLE user_sessions IS 'User sessions for JWT token management and security tracking';
COMMENT ON COLUMN user_sessions.id IS 'Primary identifier for the session';
COMMENT ON COLUMN user_sessions.user_id IS 'Foreign key to users table';
COMMENT ON COLUMN user_sessions.tenant_id IS 'Tenant identifier for multi-tenant isolation';
COMMENT ON COLUMN user_sessions.session_id IS 'Unique session identifier';
COMMENT ON COLUMN user_sessions.token_hash IS 'Hashed JWT access token for blacklisting';
COMMENT ON COLUMN user_sessions.refresh_token_hash IS 'Hashed refresh token for security';
COMMENT ON COLUMN user_sessions.device_info IS 'Device fingerprinting data in JSON format';
COMMENT ON COLUMN user_sessions.ip_address IS 'Client IP address for security tracking';
COMMENT ON COLUMN user_sessions.user_agent IS 'Browser/client identification string';
COMMENT ON COLUMN user_sessions.expires_at IS 'Token expiration timestamp';
COMMENT ON COLUMN user_sessions.last_activity IS 'Last activity timestamp for session management';
COMMENT ON COLUMN user_sessions.is_active IS 'Session status flag';
COMMENT ON COLUMN user_sessions.created_at IS 'Session creation timestamp';
COMMENT ON COLUMN user_sessions.updated_at IS 'Last update timestamp';