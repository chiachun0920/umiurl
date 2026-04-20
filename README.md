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

## API Contract

The Swagger/OpenAPI file for F2E integration is available at:

```sh
docs/openapi.yaml
```

Check that the file exists with:

```sh
make swagger
```

## Development

```sh
go mod tidy
go test ./...
go test -race ./...
go build ./...
```
