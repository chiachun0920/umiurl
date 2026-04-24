# umiurl

Short URL service for social sharing. It supports short URL creation, crawler-aware Open Graph previews, redirects, referral tracking, click analytics, and simple conversion attribution.

## Run Locally

Start Postgres:

```sh
docker compose up -d
```

Apply the schema:

```sh
psql "postgres://umiurl:umiurl@localhost:5433/umiurl?sslmode=disable" -f migrations/001_init.sql
```

Run the API:

```sh
export DATABASE_URL="postgres://umiurl:umiurl@localhost:5433/umiurl?sslmode=disable"
export APP_BASE_URL="http://localhost:8080"
go run ./cmd/api
```

## Example Requests

Create a short URL:

```sh
curl -X POST http://localhost:8080/urls \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/article","referral_code":"user_001","campaign":"spring"}'
```

Follow a short URL:

```sh
curl -i http://localhost:8080/{code}
```

Record a conversion:

```sh
curl -X POST http://localhost:8080/conversions \
  -H 'Content-Type: application/json' \
  -d '{"code":"{code}","value":99.00}'
```

View analytics:

```sh
curl http://localhost:8080/urls/{code}/analytics
```

## API Usage Scenarios

### 1. User creates a short URL

The frontend sends the original URL to the API. The API fetches preview metadata, creates a 7-character short code, stores referral/campaign data, and returns the short URL.

```sh
curl -X POST http://localhost:8080/urls \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://www.youtube.com/watch?v=BkszA-MvjXA","referral_code":"user_001","campaign":"launch"}'
```

Use `short_url` from the response as the shareable link.

### 2. User opens the short URL

A normal browser visit redirects to the original URL with `302`. The API records a click event for analytics.

```sh
curl -i http://localhost:8080/{code}
```

### 3. Social platform crawls the short URL

When Facebook, Threads, Slack, or another crawler visits the short URL, the API returns HTML with Open Graph tags instead of redirecting. This lets the platform show title, description, and image preview.

```sh
curl -H 'User-Agent: facebookexternalhit/1.1' http://localhost:8080/{code}
```

### 4. User completes a conversion

After the user performs a target action, such as signup or purchase, the frontend/backend records a conversion with the short code.

```sh
curl -X POST http://localhost:8080/conversions \
  -H 'Content-Type: application/json' \
  -d '{"code":"{code}","value":99.00}'
```

If `referral_code` or `campaign` is omitted, the API uses the values stored on the short URL.

### 5. Dashboard reads analytics

The frontend dashboard can fetch click and conversion totals plus breakdowns by referral, campaign, platform, device, and country.

```sh
curl http://localhost:8080/urls/{code}/analytics
```

## API Contract

The Swagger/OpenAPI file for F2E integration is available at:

```sh
docs/openapi.yaml
```

Check that the file exists with:

```sh
make swagger
```

## Contributor Docs

New contributors should start with [Architecture and Business Guide](docs/architecture.md). For deeper details, read [Data Models](docs/data-models.md), [Database Tables](docs/database-tables.md), and [Usecase Flows](docs/usecase-flows.md).

## Assignment Checklist

- Short URL generation and redirect: `POST /urls`, `GET /{code}`
- Social link preview: crawler-aware `GET /{code}` returns OG HTML
- Referral tracking: `referral_code` and `campaign` are stored on short URLs and copied to click/conversion events
- Analytics and attribution: `GET /urls/{code}/analytics`
- Conversion attribution: `POST /conversions`
- Local runnable setup: Docker Compose, migration, Makefile, and README commands

## Development

```sh
go mod tidy
go test ./...
```
