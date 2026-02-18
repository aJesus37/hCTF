-- Add team_id to hint_unlocks to track which team the user was in when unlocking
ALTER TABLE hint_unlocks ADD COLUMN team_id TEXT REFERENCES teams(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_hint_unlocks_team ON hint_unlocks(team_id);
