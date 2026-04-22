package integration

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jackc/pgx/v5"
)

const resetSchemaSQL = `
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
`

func resetDb() {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, databaseURL())
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

	conn, err := pgx.Connect(ctx, databaseURL())
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
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		return databaseURL
	}

	return "postgres://umiurl:umiurl@localhost:5433/umiurl?sslmode=disable"
}

func migrationPath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("resolve setupdb.go path")
	}

	return filepath.Join(filepath.Dir(file), "..", "..", "migrations", "001_init.sql")
}
