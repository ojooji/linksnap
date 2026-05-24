CREATE TABLE IF NOT EXISTS clicks (
    id         BIGSERIAL PRIMARY KEY,
    url_id     BIGINT NOT NULL REFERENCES urls (id) ON DELETE CASCADE,
    ip         INET,
    user_agent TEXT,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_clicks_url_id ON clicks (url_id);
