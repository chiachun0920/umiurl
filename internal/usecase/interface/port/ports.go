package port

import (
	"context"
	"time"

	"umiurl/internal/domain"
)

type CodeGenerator interface {
	Generate() (string, error)
}

type Clock interface {
	Now() time.Time
}

type MetadataFetcher interface {
	Fetch(ctx context.Context, rawURL string) (domain.PreviewMetadata, error)
}

type RequestClassifier interface {
	IsCrawler(userAgent string) bool
	Platform(userAgent, referer string) string
	Device(userAgent string) string
	Country(headers map[string]string) string
}
