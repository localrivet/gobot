-- MCP Session queries for organization selection persistence

-- name: GetMCPSession :one
SELECT * FROM mcp_sessions
WHERE session_id = ?;

-- name: GetMCPSessionByUser :one
-- Fallback: get most recent org selection for a user (when session ID changes)
SELECT * FROM mcp_sessions
WHERE user_id = ? AND org_id IS NOT NULL
ORDER BY updated_at DESC
LIMIT 1;

-- name: UpsertMCPSession :exec
-- Persist org selection (upsert to handle both new and existing sessions)
INSERT INTO mcp_sessions (session_id, user_id, org_id, updated_at)
VALUES (?, ?, ?, unixepoch())
ON CONFLICT (session_id) DO UPDATE SET
    org_id = excluded.org_id,
    updated_at = unixepoch();

-- name: DeleteMCPSession :exec
DELETE FROM mcp_sessions WHERE session_id = ?;

-- name: CleanupOldMCPSessions :exec
-- Run periodically to clean up stale sessions (older than 7 days)
DELETE FROM mcp_sessions
WHERE updated_at < unixepoch() - 604800;
