CREATE TABLE IF NOT EXISTS urls (
    id         BIGSERIAL PRIMARY KEY,
    code       VARCHAR(10) UNIQUE NOT NULL,
    original   TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_urls_code ON urls (code);
