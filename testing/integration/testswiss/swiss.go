package testswiss

import (
	"context"
	"log"
	"os"
	"time"
	"umiurl/internal/adapter/gateway"
	"umiurl/internal/adapter/repository"
	"umiurl/internal/controller"
	"umiurl/internal/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

var swiss *TestSwiss

type TestSwiss struct {
	Controller *controller.Controller
	EchoServer *echo.Echo
	TearDown   func()
}

func NewPool() *pgxpool.Pool {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("TEST_DATABASE_URL is required")
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	return pool
}
func NewTestSwiss() *TestSwiss {
	pool := NewPool()
	repo := repository.NewPostgresRepository(pool)
	service := usecase.NewService(usecase.NewServiceInput{
		ShortURLs: repo,
		Metadata:  NewMockMetadataFetcher(),
		Codes:     gateway.NewBase62CodeGenerator(7),
		Clock:     gateway.RealClock{},
	})

	e := echo.New()
	controller.New(service).Register(e)
	swiss := &TestSwiss{
		Controller: controller.New(service),
		EchoServer: e,
		TearDown: func() {
			pool.Close()
		},
	}
	return swiss
}
