-- Forward only. Analytics metadata keyed by the existing safe effective URL
-- identity; deliberately unrelated to pricing channels, accounts and ledgers.
CREATE TABLE IF NOT EXISTS upstream_url_channel_remarks (
    upstream TEXT PRIMARY KEY CHECK (octet_length(upstream) BETWEEN 1 AND 2048),
    remark TEXT NOT NULL DEFAULT '' CHECK (char_length(remark) <= 500),
    revision BIGINT NOT NULL CHECK (revision > 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- Empty remarks remain as versioned tombstones so a stale editor cannot
-- recreate a cleared remark at revision zero (ABA).
COMMENT ON TABLE upstream_url_channel_remarks IS 'Admin analytics notes for safe normalized effective URL identities; never routing or pricing configuration';
