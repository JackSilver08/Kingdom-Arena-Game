CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username    VARCHAR(30) UNIQUE NOT NULL,
    email       VARCHAR(255) UNIQUE NOT NULL,
    password    VARCHAR(255) NOT NULL,
    elo_rating  INT DEFAULT 1000,
    wins        INT DEFAULT 0,
    losses      INT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);
