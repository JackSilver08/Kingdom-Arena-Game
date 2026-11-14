CREATE TABLE IF NOT EXISTS matches (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player1_id  UUID REFERENCES users(id),
    player2_id  UUID REFERENCES users(id),
    winner_id   UUID REFERENCES users(id),
    duration_s  INT,
    end_reason  VARCHAR(20),
    played_at   TIMESTAMPTZ DEFAULT NOW()
);
