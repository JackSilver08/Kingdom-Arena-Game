CREATE TABLE IF NOT EXISTS leaderboard (
    user_id     UUID PRIMARY KEY REFERENCES users(id),
    elo_rating  INT NOT NULL,
    rank        INT,
    season      INT DEFAULT 1,
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);
