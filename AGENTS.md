# Repository Guidelines

## Project Structure & Module Organization

This is a Go module named `umiurl` for a short URL service. The root contains:

- `go.mod`: module declaration and Go version.
- `assignment.txt`: product and system requirements.
- `tags`: generated ctags index.

As code is added, prefer a clean architecture layout under `internal/`:

- `cmd/<service>/main.go` for binaries, e.g. `cmd/api/main.go`.
- `internal/controller/` for HTTP handlers, request parsing, and responses.
- `internal/usecase/` for workflows such as creating URLs, redirects, and analytics.
- `internal/usecase/interface/port/xxx.go` for usecase-owned ports: clocks, ID generators, and gateways.
- `internal/usecase/interface/repository/xxx.go` for repository contracts.
- `internal/domain/` for entities, value objects, services, and framework-free business rules.
- `internal/adapter/repository/` for database implementations of repository interfaces.
- `internal/adapter/gateway/` for external services such as metadata, geo-IP, or analytics exports.
- `pkg/registry/registry.go` for `NewController`, the composition root that creates controllers, usecases, repositories, gateways, and components.
- `pkg/` otherwise only for externally reusable packages.
- `testdata/` beside tests.

## Build, Test, and Development Commands

- `go mod tidy`: sync dependencies after imports change.
- `go test ./...`: run all tests.
- `go test -race ./...`: run race-detected tests.
- `go run ./cmd/api`: run the API once the binary exists.
- `go build ./...`: compile all packages.

Keep these commands working from the repository root.

## Coding Style & Naming Conventions

Use standard Go formatting. Run `gofmt` or `go fmt ./...` before committing. Package names should be short, lowercase, and singular where natural, such as `controller`, `usecase`, or `domain`. Exported identifiers need clear names and doc comments when public. Prefer dependency injection through usecase interfaces; wire concretes only in `pkg/registry`.

## Testing Guidelines

Use Go's built-in `testing` package unless the repository adopts another framework. Name test files `*_test.go` and test functions `TestXxx`. Favor table-driven tests for URL generation, redirects, attribution, and metadata parsing. Benchmark latency-sensitive lookup paths. Store fixtures in `testdata/`.

## Commit & Pull Request Guidelines

This repository has no commit history yet, so no local convention is established. Use concise imperative commit messages, for example `Add short URL generator`. Pull requests should include a summary, tests run, and notes for new environment variables, migrations, or setup. Include sample requests/responses when changing API behavior.

## Security & Configuration Tips

Do not commit secrets, API keys, production URLs, or private tokens. Keep environment-specific settings in `.env` or shell config, and document required variables in `.env.example` when introduced. Validate and normalize user-submitted URLs before storing or redirecting.
