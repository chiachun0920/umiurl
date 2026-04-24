# System Design

This document maps the assignment requirements to the current implementation and explains how the service should evolve toward a production-ready design.

## Requirement Mapping

### 1. Short URL generation and redirect

Current implementation:

- `POST /urls` creates a short URL from a long URL.
- `GET /{code}` resolves the short code and redirects normal users to the original URL.
- Short codes are random 7-character base62 strings with Postgres uniqueness enforcement and retry-on-collision logic.

Production evolution:

- Run multiple stateless API replicas behind a load balancer.
- Add a cache layer for `code -> original_url` lookup to reduce database reads on heavy redirect traffic.

### 2. Link preview

Current implementation:

- Metadata is fetched on create and cached in `short_urls`.
- When a crawler visits `GET /{code}`, the service refreshes OG metadata and returns preview HTML instead of redirecting.

Production evolution:

- Move metadata refresh to an async job queue for slow upstream sites.
- Use background refresh plus stale cache fallback to avoid blocking crawler traffic on long fetches.

### 3. Referral tracking

Current implementation:

- `referral_code` and `campaign` are accepted on create.
- Those values are stored on `short_urls` and copied into click/conversion events for stable historical attribution.

Production evolution:

- Allow richer attribution context such as source medium, content ID, or creator ID.
- Consider signed/share tokens if referral authenticity becomes important.

### 4. Analytics and attribution

Current implementation:

- Normal redirect visits create `click_events` with platform, device, country, referral, and campaign.
- `POST /conversions` records attributed `conversion_events`.
- `GET /urls/{code}/analytics` returns click/conversion totals and grouped breakdowns.

Production evolution:

- Push click/conversion writes through an event queue during traffic spikes.
- Add scheduled aggregation or materialized views for large-scale analytics queries.
- Extend analytics to include conversion value totals and ROI per campaign.

## Non-Functional Design

### High Availability

Current local implementation is a single Echo process with one Postgres instance, which is acceptable for local demo but not sufficient for true HA.

Target production design:

- Run multiple stateless API replicas behind a load balancer.
- Use health checks and rolling deploys so redirect traffic is not interrupted during releases.
- Use managed Postgres or replicated Postgres with failover.
- Store no request-critical session state in memory.
- Treat click recording and preview refresh as non-blocking where possible so core redirect remains available.

### Low Latency

Current latency choices:

- Redirect path does not fetch metadata for normal users.
- Click recording is best-effort and does not block redirect behavior.
- Create path currently fetches metadata synchronously for better preview quality.

Tradeoff:

- Synchronous metadata fetch can slow `POST /urls` if the source page is slow.

Target production improvements:

- Return a short URL quickly with fallback preview, then refresh metadata asynchronously.
- Cache resolved short codes in memory or Redis.
- Tune database indexes and connection pooling for redirect-heavy workloads.

### Scalability

Current scalability baseline:

- Clean architecture isolates controllers, usecases, repositories, and gateways.
- Event tables separate click/conversion data from short URL records.
- Analytics queries already use indexes on short URL, referral, and campaign columns.

Target production improvements:

- Add Redis or CDN caching for hot redirect lookups.
- Partition or archive event tables as traffic grows.
- Send click/conversion writes to a queue for burst handling.
- Add pre-aggregated analytics tables for dashboards instead of scanning raw event tables for every request.

## Local Implementation vs Production Evolution

The current repository is intentionally optimized for assignment delivery:

- local Postgres with Docker Compose,
- simple synchronous write path,
- direct Postgres aggregation,
- OpenAPI and contributor docs,
- runnable API with tests.

This is enough to demonstrate the required product behavior locally. The production design above explains how the same architecture can evolve toward HA, lower latency, and higher scalability without changing the public API shape.
