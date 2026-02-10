-- Recreate teams table without new columns
DROP INDEX IF EXISTS idx_teams_invite_id;

ALTER TABLE teams RENAME TO teams_new;

CREATE TABLE teams (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    owner_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Copy data back
INSERT INTO teams (id, name, description, owner_id, created_at, updated_at)
SELECT id, name, description, owner_id, created_at, updated_at
FROM teams_new;

DROP TABLE teams_new;
