-- +goose Up
-- Organizations and Team Management

-- =============================================================================
-- ORGANIZATIONS
-- =============================================================================

CREATE TABLE IF NOT EXISTS organizations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE COLLATE NOCASE,
    logo_url TEXT,
    owner_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_organizations_slug ON organizations(slug);
CREATE INDEX IF NOT EXISTS idx_organizations_owner ON organizations(owner_id);

-- =============================================================================
-- ORGANIZATION MEMBERS
-- =============================================================================

CREATE TABLE IF NOT EXISTS organization_members (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member', -- owner, admin, member
    joined_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    UNIQUE(organization_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_org_members_org ON organization_members(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_members_user ON organization_members(user_id);

-- =============================================================================
-- ORGANIZATION INVITES
-- =============================================================================

CREATE TABLE IF NOT EXISTS organization_invites (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email TEXT NOT NULL COLLATE NOCASE,
    role TEXT NOT NULL DEFAULT 'member',
    token TEXT NOT NULL UNIQUE,
    invited_by TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    UNIQUE(organization_id, email)
);

CREATE INDEX IF NOT EXISTS idx_org_invites_token ON organization_invites(token);
CREATE INDEX IF NOT EXISTS idx_org_invites_email ON organization_invites(email);
CREATE INDEX IF NOT EXISTS idx_org_invites_org ON organization_invites(organization_id);

-- =============================================================================
-- ADD CURRENT ORG TO USER PREFERENCES
-- =============================================================================

ALTER TABLE user_preferences ADD COLUMN current_organization_id TEXT REFERENCES organizations(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE user_preferences DROP COLUMN current_organization_id;

DROP INDEX IF EXISTS idx_org_invites_org;
DROP INDEX IF EXISTS idx_org_invites_email;
DROP INDEX IF EXISTS idx_org_invites_token;
DROP TABLE IF EXISTS organization_invites;

DROP INDEX IF EXISTS idx_org_members_user;
DROP INDEX IF EXISTS idx_org_members_org;
DROP TABLE IF EXISTS organization_members;

DROP INDEX IF EXISTS idx_organizations_owner;
DROP INDEX IF EXISTS idx_organizations_slug;
DROP TABLE IF EXISTS organizations;
