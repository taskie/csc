-- +migrate Up
CREATE TABLE IF NOT EXISTS managements (
    id INTEGER PRIMARY KEY,
    base_path TEXT NOT NULL,
    type TEXT NOT NULL,
    mtime DATETIME NOT NULL,
    status TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS managements;
