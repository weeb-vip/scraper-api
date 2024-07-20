CREATE TABLE IF NOT EXISTS thetvdb_link (
    id VARCHAR(36) PRIMARY KEY,
    anime_id VARCHAR(36) NOT NULL,
    thetvdb_id VARCHAR(60) NOT NULL,
);