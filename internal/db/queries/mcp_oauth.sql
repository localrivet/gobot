-- MCP OAuth queries for Dynamic Client Registration and OAuth flows

-- =============================================================================
-- OAuth Clients
-- =============================================================================

-- name: CreateMCPOAuthClient :one
INSERT INTO mcp_oauth_clients (id, client_id, client_secret_hash, name, description, redirect_uris, scopes, is_confidential)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetMCPOAuthClientByClientID :one
SELECT * FROM mcp_oauth_clients
WHERE client_id = ?;

-- name: GetMCPOAuthClientByID :one
SELECT * FROM mcp_oauth_clients
WHERE id = ?;

-- =============================================================================
-- OAuth Authorization Codes
-- =============================================================================

-- name: CreateMCPOAuthCode :one
INSERT INTO mcp_oauth_codes (id, client_id, user_id, code_hash, redirect_uri, scopes, code_challenge, code_challenge_method, expires_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetMCPOAuthCodeByHash :one
SELECT
    c.*,
    cl.client_id as oauth_client_id
FROM mcp_oauth_codes c
JOIN mcp_oauth_clients cl ON c.client_id = cl.id
WHERE c.code_hash = ? AND c.used_at IS NULL;

-- name: MarkMCPOAuthCodeUsed :exec
UPDATE mcp_oauth_codes
SET used_at = datetime('now')
WHERE id = ?;

-- =============================================================================
-- OAuth Tokens
-- =============================================================================

-- name: CreateMCPOAuthToken :one
INSERT INTO mcp_oauth_tokens (id, client_id, user_id, access_token_hash, refresh_token_hash, scopes, expires_at, refresh_expires_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetMCPOAuthTokenByAccessHash :one
SELECT * FROM mcp_oauth_tokens
WHERE access_token_hash = ? AND revoked_at IS NULL;

-- name: GetMCPOAuthTokenByRefreshHash :one
SELECT * FROM mcp_oauth_tokens
WHERE refresh_token_hash = ? AND revoked_at IS NULL;

-- name: RevokeMCPOAuthToken :exec
UPDATE mcp_oauth_tokens
SET revoked_at = datetime('now')
WHERE id = ?;

-- name: RevokeUserMCPOAuthTokens :exec
UPDATE mcp_oauth_tokens
SET revoked_at = datetime('now')
WHERE user_id = ? AND revoked_at IS NULL;
