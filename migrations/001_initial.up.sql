-- Users (authentication and profiles)
CREATE TABLE users (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    name TEXT NOT NULL,
    avatar_url TEXT,
    team_id TEXT REFERENCES teams(id) ON DELETE SET NULL,
    is_admin BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Teams (collaborative competition)
CREATE TABLE teams (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    owner_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Challenges (containers for questions)
CREATE TABLE challenges (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    name TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    category TEXT NOT NULL,
    difficulty TEXT NOT NULL,
    tags JSON,
    visible BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Questions (individual flags within challenges)
CREATE TABLE questions (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    challenge_id TEXT NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    flag TEXT NOT NULL,
    flag_mask TEXT,
    case_sensitive BOOLEAN DEFAULT 0,
    points INTEGER DEFAULT 100,
    file_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Hints (optional help for questions)
CREATE TABLE hints (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    question_id TEXT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    cost INTEGER DEFAULT 0,
    "order" INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Submissions (answer attempts)
CREATE TABLE submissions (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    question_id TEXT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id TEXT REFERENCES teams(id) ON DELETE SET NULL,
    submitted_flag TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(question_id, user_id)
);

-- Hint Unlocks (track who unlocked which hints)
CREATE TABLE hint_unlocks (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    hint_id TEXT NOT NULL REFERENCES hints(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(hint_id, user_id)
);

-- Indexes
CREATE INDEX idx_submissions_user ON submissions(user_id, created_at);
CREATE INDEX idx_submissions_team ON submissions(team_id, created_at);
CREATE INDEX idx_submissions_question ON submissions(question_id, is_correct);
CREATE INDEX idx_questions_challenge ON questions(challenge_id);
CREATE INDEX idx_challenges_category ON challenges(category, visible);
