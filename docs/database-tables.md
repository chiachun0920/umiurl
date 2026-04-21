# Database Tables

Schema is defined in [001_init.sql](../migrations/001_init.sql). Postgres is the source of truth for short URLs, click events, and conversion events.

## short_urls

Purpose: stores the canonical short link record.

Columns:

- `id BIGSERIAL PRIMARY KEY`: internal identifier used by event tables.
- `code VARCHAR(32) NOT NULL UNIQUE`: public short code. The app currently generates 7-character base62 codes; the wider column leaves room for future length changes.
- `original_url TEXT NOT NULL`: redirect target.
- `referral_code TEXT NOT NULL DEFAULT ''`: default referral identity used for click/conversion attribution.
- `campaign TEXT NOT NULL DEFAULT ''`: default campaign identity.
- `preview_title TEXT NOT NULL DEFAULT ''`: cached social preview title.
- `preview_description TEXT NOT NULL DEFAULT ''`: cached social preview description.
- `preview_image_url TEXT NOT NULL DEFAULT ''`: cached social preview image URL.
- `created_at TIMESTAMPTZ NOT NULL`: creation timestamp.
- `updated_at TIMESTAMPTZ NOT NULL`: last update timestamp.

Constraints and indexes:

- `code` is unique so the public short URL resolves to exactly one original URL.

Business notes:

- Preview columns are filled when a short URL is created and refreshed when crawlers visit.
- Referral and campaign fields are copied into event tables so historical analytics remain stable even if short URL defaults change later.

## click_events

Purpose: immutable log of normal user redirect clicks.

Columns:

- `id BIGSERIAL PRIMARY KEY`: internal event identifier.
- `short_url_id BIGINT NOT NULL REFERENCES short_urls(id) ON DELETE CASCADE`: related short URL.
- `referral_code TEXT NOT NULL DEFAULT ''`: referral attributed at click time.
- `campaign TEXT NOT NULL DEFAULT ''`: campaign attributed at click time.
- `user_agent TEXT NOT NULL DEFAULT ''`: raw user agent.
- `platform TEXT NOT NULL DEFAULT ''`: classified source platform.
- `device TEXT NOT NULL DEFAULT ''`: classified device.
- `country TEXT NOT NULL DEFAULT ''`: country code from request headers, or `unknown`.
- `occurred_at TIMESTAMPTZ NOT NULL`: click timestamp.

Indexes:

- `idx_click_events_short_url_id`: supports analytics lookup by short URL.
- `idx_click_events_referral_code`: supports referral breakdowns.
- `idx_click_events_campaign`: supports campaign breakdowns.

Business notes:

- Only normal redirect visits create click events.
- Crawler preview requests return HTML and do not count as clicks.
- `ON DELETE CASCADE` removes click history if the parent short URL is deleted.

## conversion_events

Purpose: immutable log of business outcomes after a short URL journey.

Columns:

- `id BIGSERIAL PRIMARY KEY`: internal event identifier.
- `short_url_id BIGINT NOT NULL REFERENCES short_urls(id) ON DELETE CASCADE`: related short URL.
- `referral_code TEXT NOT NULL DEFAULT ''`: referral attributed at conversion time.
- `campaign TEXT NOT NULL DEFAULT ''`: campaign attributed at conversion time.
- `value NUMERIC(12, 2) NOT NULL DEFAULT 0`: optional business value, such as purchase amount or estimated lead value.
- `occurred_at TIMESTAMPTZ NOT NULL`: conversion timestamp.

Indexes:

- `idx_conversion_events_short_url_id`: supports analytics lookup by short URL.
- `idx_conversion_events_referral_code`: supports referral attribution.
- `idx_conversion_events_campaign`: supports campaign attribution.

Business notes:

- Current analytics count conversions and group them by referral/campaign.
- `value` is stored for future revenue/value reporting, even though current analytics do not sum it.
- If conversion request attribution fields are omitted, the usecase uses the short URL defaults.

## Schema Change Rules

When changing schema:

- Add or update SQL in `migrations/`.
- Update domain models if the field is part of business logic.
- Update repository queries in `internal/adapter/repository`.
- Update usecase/controller tests when behavior changes.
- Update [openapi.yaml](openapi.yaml) and README when API responses or inputs change.
