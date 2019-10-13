-- +migrate Up
CREATE TABLE IF NOT EXISTS namespaces (
    name VARCHAR(100) NOT NULL,
    url VARCHAR(500) NOT NULL,
    type VARCHAR(10) NOT NULL,
    csc_db_size BIGINT NOT NULL,
    csc_db_mtime DATETIME NOT NULL,
    csc_db_sha256 CHAR(64) NOT NULL,
    status VARCHAR(10) NOT NULL,
    description VARCHAR(1000) NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    PRIMARY KEY (name)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE INDEX namespaces_csc_db_mtime ON namespaces (csc_db_mtime);
CREATE INDEX namespaces_updated_at ON namespaces (updated_at);

-- +migrate Down
DROP TABLE IF EXISTS namespaces;
