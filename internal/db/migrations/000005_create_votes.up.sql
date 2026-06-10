CREATE TABLE votes (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id     UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    option_id   UUID NOT NULL REFERENCES options(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_votes_room_user UNIQUE (room_id, user_id)
);

CREATE INDEX idx_votes_room_id   ON votes(room_id);
CREATE INDEX idx_votes_option_id ON votes(option_id);
CREATE INDEX idx_votes_user_id   ON votes(user_id);