-- +migrate Up
CREATE TABLE IF NOT EXISTS objects (
    id INTEGER AUTO_INCREMENT,
    namespace VARCHAR(100) NOT NULL,
    path VARCHAR(500) NOT NULL,
    type VARCHAR(10) NOT NULL,
    size BIGINT NOT NULL,
    mtime DATETIME NOT NULL,
    sha256 CHAR(64) NOT NULL,
    status VARCHAR(10) NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE (namespace, path),
    PRIMARY KEY (id)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE INDEX objects_path ON objects (path); 
CREATE INDEX objects_sha256_path ON objects (sha256, path);
CREATE INDEX objects_mtime ON objects (mtime);
CREATE INDEX objects_updated_at ON objects (updated_at);

-- +migrate Down
DROP TABLE IF EXISTS objects;
