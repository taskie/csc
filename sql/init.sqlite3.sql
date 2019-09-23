CREATE TABLE objects (
    id INTEGER PRIMARY KEY,
    path TEXT UNIQUE NOT NULL,
    type TEXT NOT NULL,
    size INTEGER NOT NULL,
    mtime DATETIME NOT NULL,
    sha256 TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE INDEX objects_path ON objects (path); 
CREATE INDEX objects_sha256_path ON objects (sha256, path);
CREATE INDEX objects_mtime ON objects (mtime);
CREATE INDEX objects_updated_at ON objects (updated_at);
