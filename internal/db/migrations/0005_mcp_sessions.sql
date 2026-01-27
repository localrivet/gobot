-- MCP Sessions table for persisting organization selection across server restarts
-- +goose Up
CREATE TABLE IF NOT EXISTS mcp_sessions (
    session_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    org_id TEXT,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE SET NULL
);

CREATE INDEX idx_mcp_sessions_user_id ON mcp_sessions(user_id);
CREATE INDEX idx_mcp_sessions_updated_at ON mcp_sessions(updated_at);

-- +goose Down
DROP INDEX IF EXISTS idx_mcp_sessions_updated_at;
DROP INDEX IF EXISTS idx_mcp_sessions_user_id;
DROP TABLE IF EXISTS mcp_sessions;
