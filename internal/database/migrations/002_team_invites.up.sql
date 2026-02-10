-- Recreate teams table with new columns
ALTER TABLE teams RENAME TO teams_old;

CREATE TABLE teams (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    owner_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    invite_id TEXT UNIQUE NOT NULL,
    invite_permission TEXT NOT NULL DEFAULT 'owner_only',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Copy data and generate unique invite codes
INSERT INTO teams (id, name, description, owner_id, invite_id, invite_permission, created_at, updated_at)
SELECT id, name, description, owner_id, lower(hex(randomblob(16))), 'owner_only', created_at, updated_at
FROM teams_old;

DROP TABLE teams_old;

-- Create index for fast invite code lookups
CREATE INDEX idx_teams_invite_id ON teams(invite_id);
