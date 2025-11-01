-- Migration: Create activity_logs table
-- Created: Authentication Service
-- Description: Creates the activity_logs table for audit trail and security monitoring

-- Create activity_logs table
CREATE TABLE IF NOT EXISTS activity_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID,
    tenant_id UUID NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id UUID,
    old_values JSON,
    new_values JSON,
    ip_address VARCHAR(45),
    user_agent TEXT,
    session_id VARCHAR(255),
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    context JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance and querying
CREATE INDEX idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_tenant_id ON activity_logs(tenant_id);
CREATE INDEX idx_activity_logs_action ON activity_logs(action);
CREATE INDEX idx_activity_logs_resource_type ON activity_logs(resource_type);
CREATE INDEX idx_activity_logs_resource_id ON activity_logs(resource_id);
CREATE INDEX idx_activity_logs_session_id ON activity_logs(session_id);
CREATE INDEX idx_activity_logs_success ON activity_logs(success);
CREATE INDEX idx_activity_logs_created_at ON activity_logs(created_at);

-- Create composite indexes for common queries
CREATE INDEX idx_activity_logs_tenant_action ON activity_logs(tenant_id, action);
CREATE INDEX idx_activity_logs_user_action ON activity_logs(user_id, action) WHERE user_id IS NOT NULL;
CREATE INDEX idx_activity_logs_resource ON activity_logs(resource_type, resource_id) WHERE resource_id IS NOT NULL;

-- Add foreign key constraints (SET NULL to preserve audit trail)
ALTER TABLE activity_logs
    ADD CONSTRAINT fk_activity_logs_user_id
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- Add table comments for documentation
COMMENT ON TABLE activity_logs IS 'Activity log entries for audit trail and security monitoring';
COMMENT ON COLUMN activity_logs.id IS 'Primary identifier for the activity log entry';
COMMENT ON COLUMN activity_logs.user_id IS 'Foreign key to users table (nullable for system activities)';
COMMENT ON COLUMN activity_logs.tenant_id IS 'Tenant identifier for multi-tenant isolation';
COMMENT ON COLUMN activity_logs.action IS 'Action performed (e.g., login, logout, create, update, delete)';
COMMENT ON COLUMN activity_logs.resource_type IS 'Type of resource affected (e.g., user, product, invoice)';
COMMENT ON COLUMN activity_logs.resource_id IS 'ID of affected resource';
COMMENT ON COLUMN activity_logs.old_values IS 'Previous state for audit in JSON format';
COMMENT ON COLUMN activity_logs.new_values IS 'New state for audit in JSON format';
COMMENT ON COLUMN activity_logs.ip_address IS 'Client IP address for security tracking';
COMMENT ON COLUMN activity_logs.user_agent IS 'Browser/client identification string';
COMMENT ON COLUMN activity_logs.session_id IS 'Related session identifier';
COMMENT ON COLUMN activity_logs.success IS 'Action success status';
COMMENT ON COLUMN activity_logs.error_message IS 'Error details if action failed';
COMMENT ON COLUMN activity_logs.context IS 'Additional context data in JSON format';
COMMENT ON COLUMN activity_logs.created_at IS 'Activity timestamp';

-- Create a partitioning strategy for large-scale deployments (optional)
-- This would be implemented in production environments with high volume
-- Uncomment and modify as needed for your specific requirements

/*
-- Example partitioning by month for high-volume systems
CREATE TABLE activity_logs_y2024m01 PARTITION OF activity_logs
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE activity_logs_y2024m02 PARTITION OF activity_logs
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
*/