-- name: CreateOrganization :one
INSERT INTO organizations (id, name, slug, logo_url, owner_id, created_at, updated_at)
VALUES (sqlc.arg(id), sqlc.arg(name), sqlc.arg(slug), sqlc.narg(logo_url), sqlc.arg(owner_id), strftime('%s', 'now'), strftime('%s', 'now'))
RETURNING *;

-- name: GetOrganizationByID :one
SELECT * FROM organizations WHERE id = sqlc.arg(id) LIMIT 1;

-- name: GetOrganizationBySlug :one
SELECT * FROM organizations WHERE slug = sqlc.arg(slug) LIMIT 1;

-- name: ListUserOrganizations :many
SELECT o.* FROM organizations o
JOIN organization_members om ON o.id = om.organization_id
WHERE om.user_id = sqlc.arg(user_id)
ORDER BY o.created_at DESC;

-- name: UpdateOrganization :exec
UPDATE organizations
SET name = COALESCE(sqlc.narg(name), name),
    slug = COALESCE(sqlc.narg(slug), slug),
    logo_url = COALESCE(sqlc.narg(logo_url), logo_url),
    updated_at = strftime('%s', 'now')
WHERE id = sqlc.arg(id);

-- name: DeleteOrganization :exec
DELETE FROM organizations WHERE id = sqlc.arg(id);

-- name: CheckSlugExists :one
SELECT CASE WHEN COUNT(*) > 0 THEN 1 ELSE 0 END as found
FROM organizations WHERE LOWER(slug) = LOWER(sqlc.arg(slug));

-- Organization Members

-- name: AddOrganizationMember :one
INSERT INTO organization_members (id, organization_id, user_id, role, joined_at)
VALUES (sqlc.arg(id), sqlc.arg(organization_id), sqlc.arg(user_id), sqlc.arg(role), strftime('%s', 'now'))
RETURNING *;

-- name: GetOrganizationMember :one
SELECT om.*, u.email, u.name as user_name, u.avatar_url
FROM organization_members om
JOIN users u ON om.user_id = u.id
WHERE om.organization_id = sqlc.arg(organization_id) AND om.user_id = sqlc.arg(user_id)
LIMIT 1;

-- name: ListOrganizationMembers :many
SELECT om.*, u.email, u.name as user_name, u.avatar_url
FROM organization_members om
JOIN users u ON om.user_id = u.id
WHERE om.organization_id = sqlc.arg(organization_id)
ORDER BY om.joined_at ASC;

-- name: UpdateMemberRole :exec
UPDATE organization_members
SET role = sqlc.arg(role)
WHERE organization_id = sqlc.arg(organization_id) AND user_id = sqlc.arg(user_id);

-- name: RemoveOrganizationMember :exec
DELETE FROM organization_members
WHERE organization_id = sqlc.arg(organization_id) AND user_id = sqlc.arg(user_id);

-- name: CountOrganizationMembers :one
SELECT COUNT(*) as count FROM organization_members WHERE organization_id = sqlc.arg(organization_id);

-- Organization Invites

-- name: CreateOrganizationInvite :one
INSERT INTO organization_invites (id, organization_id, email, role, token, invited_by, expires_at, created_at)
VALUES (sqlc.arg(id), sqlc.arg(organization_id), sqlc.arg(email), sqlc.arg(role), sqlc.arg(token), sqlc.arg(invited_by), sqlc.arg(expires_at), strftime('%s', 'now'))
RETURNING *;

-- name: GetInviteByToken :one
SELECT oi.*, o.name as organization_name, o.slug as organization_slug
FROM organization_invites oi
JOIN organizations o ON oi.organization_id = o.id
WHERE oi.token = sqlc.arg(token) AND oi.expires_at > strftime('%s', 'now')
LIMIT 1;

-- name: GetInviteByEmail :one
SELECT * FROM organization_invites
WHERE organization_id = sqlc.arg(organization_id) AND LOWER(email) = LOWER(sqlc.arg(email))
LIMIT 1;

-- name: ListOrganizationInvites :many
SELECT oi.*, u.name as inviter_name, u.email as inviter_email
FROM organization_invites oi
JOIN users u ON oi.invited_by = u.id
WHERE oi.organization_id = sqlc.arg(organization_id) AND oi.expires_at > strftime('%s', 'now')
ORDER BY oi.created_at DESC;

-- name: DeleteInvite :exec
DELETE FROM organization_invites WHERE id = sqlc.arg(id);

-- name: DeleteInviteByToken :exec
DELETE FROM organization_invites WHERE token = sqlc.arg(token);

-- name: DeleteExpiredInvites :exec
DELETE FROM organization_invites WHERE expires_at <= strftime('%s', 'now');

-- User Preferences (org related)

-- name: SetCurrentOrganization :exec
UPDATE user_preferences
SET current_organization_id = sqlc.narg(organization_id),
    updated_at = strftime('%s', 'now')
WHERE user_id = sqlc.arg(user_id);

-- name: GetCurrentOrganization :one
SELECT current_organization_id FROM user_preferences WHERE user_id = sqlc.arg(user_id) LIMIT 1;
