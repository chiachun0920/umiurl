# Usecase Flows

This document explains the main workflows implemented in [service.go](../internal/usecase/service.go). Controllers map HTTP requests into these usecases; repositories and ports provide storage and external integrations.

## CreateShortURL

Endpoint: `POST /urls`

Input:

- `url`: required original URL.
- `referral_code`: optional default referral attribution.
- `campaign`: optional default campaign attribution.

Flow:

1. Validate that `url` is an absolute `http` or `https` URL.
2. Fetch preview metadata from the original URL.
3. If metadata fetch fails, use fallback preview: title `Shared link`, description as original URL.
4. Generate a random 7-character base62 code.
5. Build a `ShortURL` domain model with trimmed referral/campaign values.
6. Insert the record through `ShortURLRepository.Create`.
7. If code collision occurs, retry up to `DefaultCollisionRetries`.
8. Return the public short URL and stored entity.

Side effects:

- Inserts one row into `short_urls`.
- May make one outbound HTTP request to resolve preview metadata.

Errors:

- Invalid URL returns `ErrInvalidURL`.
- Exhausted code collisions return `ErrCodeCollision`.
- Repository failures are returned to the controller.

Extension points:

- Change code generation through `CodeGenerator`.
- Change metadata behavior through `MetadataFetcher`.
- Add validation in the usecase, not in the controller.

## Resolve

Endpoint: `GET /{code}`

Input:

- `code`: short code path parameter.
- `User-Agent`, `Referer`, and country-related headers.

Flow:

1. Load the short URL by code.
2. Return `ErrShortURLNotFound` if no record exists.
3. Classify whether the request is from a crawler.
4. If crawler:
   - Fetch preview metadata from the original URL.
   - If fetch succeeds, merge fallback fields and update cached preview metadata.
   - Return output marked as preview.
5. If normal user:
   - Classify platform, device, and country.
   - Record a `ClickEvent` best-effort.
   - Return output for redirect.

Side effects:

- Crawler requests may update preview fields on `short_urls`.
- Normal user requests may insert one row into `click_events`.

Errors:

- Unknown code returns `ErrShortURLNotFound`.
- Metadata refresh failure does not fail crawler preview.
- Click recording failure does not block redirect.

Extension points:

- Add crawler/platform/device rules in `RequestClassifier`.
- Add geo-IP integration behind a port rather than in the controller.

## RecordConversion

Endpoint: `POST /conversions`

Input:

- `code`: required short code.
- `referral_code`: optional attribution override.
- `campaign`: optional attribution override.
- `value`: optional business value.

Flow:

1. Load short URL by code.
2. Return `ErrShortURLNotFound` if no record exists.
3. Trim request referral/campaign values.
4. If request referral/campaign are empty, use stored values from the short URL.
5. Build a `ConversionEvent`.
6. Store it through `ConversionRepository.RecordConversion`.

Side effects:

- Inserts one row into `conversion_events`.

Errors:

- Unknown code returns `ErrShortURLNotFound`.
- Repository failures are returned to the controller.

Business notes:

- A conversion is a target business action such as signup, purchase, install, booking, or form submit.
- `value` can represent order amount or estimated lead value. Current analytics count conversions but do not sum value.

Extension points:

- Add conversion value analytics in repository aggregation.
- Add event metadata by extending table, domain model, OpenAPI schema, and controller request.

## Analytics

Endpoint: `GET /urls/{code}/analytics`

Input:

- `code`: short code path parameter.

Flow:

1. Ask `AnalyticsRepository.Summary` for aggregate data.
2. Return `ErrShortURLNotFound` when the repository reports no matching short URL.
3. Return `AnalyticsSummary` to the controller.

Side effects:

- None. This is a read-only usecase.

Output:

- Total clicks.
- Total conversions.
- Click breakdowns by referral, campaign, platform, device, and country.
- Conversion breakdowns by referral and campaign.

Extension points:

- Add new breakdowns by updating repository aggregation and `AnalyticsSummary`.
- Update OpenAPI and dashboard expectations when response fields change.

## Shared Rules

- Controllers should translate HTTP input/output only.
- Usecases own business decisions and fallback behavior.
- Repository interfaces belong to `internal/usecase/interface/repository`.
- Utility/external service interfaces belong to `internal/usecase/interface/port`.
- Concrete implementations are wired only in `pkg/registry`.
