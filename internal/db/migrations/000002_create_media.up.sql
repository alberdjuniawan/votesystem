CREATE TABLE media (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    uploader_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename      VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    mime_type     VARCHAR(100) NOT NULL,
    size_bytes    BIGINT NOT NULL,
    storage_path  TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_media_uploader_id ON media(uploader_id);