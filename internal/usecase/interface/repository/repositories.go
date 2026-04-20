package repository

import (
	"context"
	"errors"
	"time"

	"umiurl/internal/domain"
)

var ErrDuplicateCode = errors.New("duplicate short code")

type ShortURLRepository interface {
	Create(ctx context.Context, shortURL domain.ShortURL) (domain.ShortURL, error)
	GetByCode(ctx context.Context, code string) (domain.ShortURL, bool, error)
	UpdatePreview(ctx context.Context, id int64, preview domain.PreviewMetadata, updatedAt time.Time) error
}

type ClickRepository interface {
	Record(ctx context.Context, event domain.ClickEvent) error
}

type ConversionRepository interface {
	RecordConversion(ctx context.Context, event domain.ConversionEvent) error
}

type AnalyticsRepository interface {
	Summary(ctx context.Context, code string) (domain.AnalyticsSummary, bool, error)
}
