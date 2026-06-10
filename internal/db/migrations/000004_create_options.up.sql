CREATE TABLE options (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id     UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    label       VARCHAR(255) NOT NULL,
    description TEXT,
    metadata    JSONB,
    media_id    UUID REFERENCES media(id) ON DELETE SET NULL,
    order_num   INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_options_room_id  ON options(room_id);
CREATE INDEX idx_options_metadata ON options USING GIN(metadata);