package registry

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"umiurl/internal/adapter/gateway"
	pgrepository "umiurl/internal/adapter/repository"
	"umiurl/internal/controller"
	"umiurl/internal/usecase"
)

func NewController(pool *pgxpool.Pool, baseURL string) *controller.Controller {
	repo := pgrepository.NewPostgresRepository(pool)
	service := usecase.NewService(usecase.NewServiceInput{
		ShortURLs:   repo,
		Clicks:      repo,
		Conversions: repo,
		Analytics:   repo,
		Codes:       gateway.NewBase62CodeGenerator(7),
		Clock:       gateway.RealClock{},
		Metadata:    gateway.NewHTTPMetadataFetcher(),
		Classifier:  gateway.SimpleRequestClassifier{},
		BaseURL:     baseURL,
		MaxRetries:  usecase.DefaultCollisionRetries,
	})
	return controller.New(service)
}
