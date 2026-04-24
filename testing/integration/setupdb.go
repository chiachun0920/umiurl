package integration

import (
	"context"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jackc/pgx/v5"
)

const resetSchemaSQL = `
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
`

func resetDb() {
	ctx := context.Background()
	dbURL := databaseURL()
	validateTestDatabaseURL(dbURL)

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	if _, err = conn.Exec(ctx, resetSchemaSQL); err != nil {
		log.Fatal(err)
	}

	sql, err := os.ReadFile(migrationPath())
	if err != nil {
		log.Fatal(err)
	}

	_, err = conn.Exec(ctx, string(sql))
	if err != nil {
		log.Fatal(err)
	}
}

func truncateDb() {
	ctx := context.Background()
	dbURL := databaseURL()
	validateTestDatabaseURL(dbURL)

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "TRUNCATE TABLE short_urls, click_events, conversion_events RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatal(err)
	}
}

func databaseURL() string {
	if databaseURL := os.Getenv("TEST_DATABASE_URL"); databaseURL != "" {
		return databaseURL
	}
	log.Fatal("TEST_DATABASE_URL is required for integration tests")
	return ""
}

func migrationPath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("resolve setupdb.go path")
	}

	return filepath.Join(filepath.Dir(file), "..", "..", "migrations", "001_init.sql")
}
func validateTestDatabaseURL(rawURL string) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}

	host := parsed.Hostname()
	if host != "localhost" && host != "127.0.0.1" {
		log.Fatalf("refusing to reset non-local integration database host %q", host)
	}

	dbName := strings.TrimPrefix(parsed.Path, "/")
	if dbName == "" {
		log.Fatal("TEST_DATABASE_URL must include a database name")
	}
	if !strings.Contains(strings.ToLower(dbName), "test") && !strings.Contains(strings.ToLower(dbName), "integration") {
		log.Fatalf("refusing to reset database %q; use a dedicated test or integration database name", dbName)
	}
}
