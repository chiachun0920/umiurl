CREATE TABLE IF NOT EXISTS short_urls (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(32) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    referral_code TEXT NOT NULL DEFAULT '',
    campaign TEXT NOT NULL DEFAULT '',
    preview_title TEXT NOT NULL DEFAULT '',
    preview_description TEXT NOT NULL DEFAULT '',
    preview_image_url TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS click_events (
    id BIGSERIAL PRIMARY KEY,
    short_url_id BIGINT NOT NULL REFERENCES short_urls(id) ON DELETE CASCADE,
    referral_code TEXT NOT NULL DEFAULT '',
    campaign TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    platform TEXT NOT NULL DEFAULT '',
    device TEXT NOT NULL DEFAULT '',
    country TEXT NOT NULL DEFAULT '',
    occurred_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS conversion_events (
    id BIGSERIAL PRIMARY KEY,
    short_url_id BIGINT NOT NULL REFERENCES short_urls(id) ON DELETE CASCADE,
    referral_code TEXT NOT NULL DEFAULT '',
    campaign TEXT NOT NULL DEFAULT '',
    value NUMERIC(12, 2) NOT NULL DEFAULT 0,
    occurred_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_click_events_short_url_id ON click_events(short_url_id);
CREATE INDEX IF NOT EXISTS idx_click_events_referral_code ON click_events(referral_code);
CREATE INDEX IF NOT EXISTS idx_click_events_campaign ON click_events(campaign);
CREATE INDEX IF NOT EXISTS idx_conversion_events_short_url_id ON conversion_events(short_url_id);
CREATE INDEX IF NOT EXISTS idx_conversion_events_referral_code ON conversion_events(referral_code);
CREATE INDEX IF NOT EXISTS idx_conversion_events_campaign ON conversion_events(campaign);
