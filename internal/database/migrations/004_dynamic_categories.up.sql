-- Categories table for dynamic challenge categorization
CREATE TABLE IF NOT EXISTS categories (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    name TEXT NOT NULL UNIQUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Difficulties table for dynamic difficulty levels
CREATE TABLE IF NOT EXISTS difficulties (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    name TEXT NOT NULL UNIQUE,
    color TEXT NOT NULL DEFAULT 'bg-gray-600 text-gray-100',
    text_color TEXT NOT NULL DEFAULT 'text-gray-400',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Seed default categories
INSERT OR IGNORE INTO categories (name, sort_order) VALUES
    ('web', 1),
    ('crypto', 2),
    ('pwn', 3),
    ('forensics', 4),
    ('misc', 5);

-- Seed default difficulties with color mappings
INSERT OR IGNORE INTO difficulties (name, color, text_color, sort_order) VALUES
    ('easy', 'bg-green-600 text-green-100', 'text-green-400', 1),
    ('medium', 'bg-yellow-600 text-yellow-100', 'text-yellow-400', 2),
    ('hard', 'bg-red-600 text-red-100', 'text-red-400', 3);
