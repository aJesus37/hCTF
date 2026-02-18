DROP INDEX IF EXISTS idx_hint_unlocks_team;

-- SQLite doesn't support DROP COLUMN on older versions; recreate table without team_id
CREATE TABLE hint_unlocks_old AS SELECT id, hint_id, user_id, created_at FROM hint_unlocks;
DROP TABLE hint_unlocks;
ALTER TABLE hint_unlocks_old RENAME TO hint_unlocks;
