# Architecture and Business Guide

This document helps new contributors understand the business purpose, request flows, and code organization of `umiurl`.

## Business Overview

`umiurl` is a short URL service for social sharing. It does more than redirect links: it creates rich link previews, tracks referral/campaign context, records clicks, and attributes conversions.

Key terms:

- **Short URL**: a generated 7-character base62 code that maps to one original URL.
- **Referral code**: identifies the person or partner who shared the link.
- **Campaign**: identifies a marketing or product campaign.
- **Click**: a normal user visit to a short URL that redirects to the original URL.
- **Conversion**: a business outcome after a click, such as signup, purchase, install, booking, or form submit.
- **Conversion value**: optional numeric value for a conversion, such as order amount or estimated lead value.
- **Attribution**: connecting clicks and conversions back to the referral code or campaign.

## User and API Scenarios

1. **Create short URL**: frontend calls `POST /urls` with the original URL and optional referral/campaign. The service resolves OG metadata, generates a short code, and stores the record.
2. **Open short URL**: normal users call `GET /{code}`. The service records a click and returns `302` to the original URL.
3. **Social preview**: crawlers call `GET /{code}` with crawler user agents. The service refreshes OG metadata and returns preview HTML instead of redirecting.
4. **Record conversion**: product or backend code calls `POST /conversions` after a target action. If referral/campaign are omitted, stored short URL values are used.
5. **Read analytics**: dashboard calls `GET /urls/{code}/analytics` for click and conversion totals plus referral/campaign/platform/device/country breakdowns.

See the machine-readable API contract in [openapi.yaml](openapi.yaml).

## Architecture Overview

The project follows a clean architecture layout:

- `cmd/api`: process entrypoint, Echo setup, middleware, and server start.
- `internal/controller`: Echo handlers, request parsing, response mapping, and route registration.
- `internal/usecase`: application workflows such as create URL, resolve URL, record conversion, and analytics.
- `internal/domain`: framework-free business data models.
- `internal/usecase/interface/repository`: storage contracts owned by usecases.
- `internal/usecase/interface/port`: external/utility contracts such as code generation, metadata fetching, clocks, and request classification.
- `internal/adapter/repository`: Postgres implementations of repository interfaces.
- `internal/adapter/gateway`: external or infrastructure implementations, including OG fetching, base62 generation, and request classification.
- `pkg/registry`: composition root. `NewController` wires concrete repositories, gateways, usecases, and controllers.

Request flow:

```text
HTTP request
  -> Echo middleware
  -> controller
  -> usecase
  -> repository/port interfaces
  -> adapter implementations
  -> Postgres or external web page
```

Business rules should live in `internal/usecase` or `internal/domain`, not in controllers or Postgres adapters.

## Important Design Decisions

- Short codes are random 7-character base62 strings.
- Postgres enforces unique short codes; the create usecase retries collisions.
- Metadata is resolved when a short URL is created and refreshed/cached when crawlers visit.
- Normal redirects avoid metadata fetching so the redirect path stays fast.
- Click recording is best-effort; redirect should still work if analytics recording fails.
- Conversion attribution is intentionally simple: no auth, user identity, or session tracking.
- CORS is open by default for development and configurable with `CORS_ALLOW_ORIGINS`.

## Data Model

Schema lives in [001_init.sql](../migrations/001_init.sql).

- `short_urls`: original URL, short code, referral/campaign, preview metadata, timestamps.
- `click_events`: one row per redirect click with referral/campaign, user agent, platform, device, and country.
- `conversion_events`: one row per conversion with referral/campaign and optional numeric value.

Analytics are calculated from event tables and returned by `GET /urls/{code}/analytics`.

## How To Contribute

- Add a new HTTP endpoint in `internal/controller`, then call a usecase.
- Add new business behavior in `internal/usecase`; define interfaces there when storage or external services are needed.
- Add or change core models in `internal/domain`.
- Add database operations by updating repository interfaces first, then implementing them in `internal/adapter/repository`.
- Add external integrations in `internal/adapter/gateway` behind a `port` interface.
- Wire new concrete dependencies only in `pkg/registry`.
- Update [openapi.yaml](openapi.yaml) when public API behavior changes.
- Follow repository rules in [AGENTS.md](../AGENTS.md).

Before submitting changes, run:

```sh
make test
make build
```

Use `go test -race ./...` for concurrency-sensitive changes.

## Common Change Examples

- **Add an analytics field**: update domain summary, repository aggregation, controller response tests, and OpenAPI schema.
- **Add a platform classifier**: update `SimpleRequestClassifier` and add classifier/usecase tests.
- **Improve metadata fallback**: update `HTTPMetadataFetcher` and metadata tests.
- **Change short code length**: update generator config in `pkg/registry`, docs, OpenAPI pattern/examples, and tests.
- **Add conversion metrics**: extend `conversion_events`, analytics aggregation, API response schema, and README examples.
