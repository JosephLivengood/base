-- +goose Up
CREATE TABLE seen_ips (
    ip TEXT PRIMARY KEY,
    num_visits INTEGER NOT NULL DEFAULT 1,
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_seen_ips_last_seen ON seen_ips(last_seen);

-- +goose Down
DROP TABLE seen_ips;
