DATABASE_URL ?= postgres://umiurl:umiurl@localhost:5433/umiurl?sslmode=disable
APP_BASE_URL ?=  https://b732eaa7dbc0.ngrok.app
PORT ?= 8080
CORS_ALLOW_ORIGINS ?= *
GOCACHE ?= /tmp/umiurl-gocache

.PHONY: start test build db-up migrate swagger

start:
	DATABASE_URL="$(DATABASE_URL)" APP_BASE_URL="$(APP_BASE_URL)" PORT="$(PORT)" CORS_ALLOW_ORIGINS="$(CORS_ALLOW_ORIGINS)" GOCACHE="$(GOCACHE)" go run ./cmd/api

test-integration:
	go test -v -count 1 \
		-cover -covermode=count -coverpkg=./... \
		-coverprofile=coverage.out \
		./testing/integration
	gcov2lcov -infile=coverage.out -outfile=lcov.info

test:
	go test -v -count 1 \
		-cover -covermode=count -coverpkg=./... \
		-coverprofile=coverage.out \
		./...
	gcov2lcov -infile=coverage.out -outfile=lcov.info

build:
	GOCACHE="$(GOCACHE)" go build ./...

db-up:
	docker compose up -d

migrate:
	psql "$(DATABASE_URL)" -f migrations/001_init.sql

swagger:
	test -f docs/openapi.yaml
	@echo "Swagger file: docs/openapi.yaml"
