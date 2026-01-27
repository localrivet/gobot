-- name: CreateLead :one
INSERT INTO leads (
    id, email, name, source, status, metadata, created_at
) VALUES (
    sqlc.arg(id), sqlc.arg(email), sqlc.narg(name), sqlc.narg(source),
    COALESCE(sqlc.narg(status), 'pending'), sqlc.narg(metadata),
    strftime('%s', 'now')
)
RETURNING id, email, name, source, status, metadata, created_at;

-- name: GetLeadByEmail :one
SELECT id, email, name, source, status, metadata, created_at
FROM leads
WHERE LOWER(email) = LOWER(sqlc.arg(email))
LIMIT 1;

-- name: GetLeadByID :one
SELECT id, email, name, source, status, metadata, created_at
FROM leads
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: UpdateLeadStatus :exec
UPDATE leads
SET status = sqlc.arg(status)
WHERE id = sqlc.arg(id);

-- name: ListLeads :many
SELECT id, email, name, source, status, metadata, created_at
FROM leads
WHERE (sqlc.narg(filter_status) IS NULL OR status = sqlc.narg(filter_status))
ORDER BY created_at DESC
LIMIT sqlc.arg(page_size) OFFSET sqlc.arg(page_offset);

-- name: CountLeads :one
SELECT COUNT(*) as count
FROM leads
WHERE (sqlc.narg(filter_status) IS NULL OR status = sqlc.narg(filter_status));

-- name: DeleteLead :exec
DELETE FROM leads WHERE id = sqlc.arg(id);
