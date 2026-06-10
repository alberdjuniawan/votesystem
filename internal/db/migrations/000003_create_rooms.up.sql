CREATE TYPE room_status AS ENUM ('draft', 'active', 'closed');
CREATE TYPE room_type   AS ENUM ('single_choice', 'multiple_choice');

CREATE TABLE rooms (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    type            room_type NOT NULL DEFAULT 'single_choice',
    status          room_status NOT NULL DEFAULT 'draft',
    show_realtime   BOOLEAN NOT NULL DEFAULT TRUE,
    max_votes       INT NOT NULL DEFAULT 1,
    starts_at       TIMESTAMPTZ,
    ends_at         TIMESTAMPTZ,
    share_code      VARCHAR(12) NOT NULL UNIQUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rooms_share_code ON rooms(share_code);
CREATE INDEX idx_rooms_owner_id   ON rooms(owner_id);
CREATE INDEX idx_rooms_status     ON rooms(status);