package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"umiurl/internal/domain"
	"umiurl/internal/usecase/interface/port"
	"umiurl/internal/usecase/interface/repository"
)

const DefaultCollisionRetries = 5

var (
	ErrInvalidURL       = errors.New("invalid url")
	ErrShortURLNotFound = errors.New("short url not found")
	ErrCodeCollision    = errors.New("could not allocate unique short code")
)

type Service struct {
	shortURLs   repository.ShortURLRepository
	clicks      repository.ClickRepository
	conversions repository.ConversionRepository
	analytics   repository.AnalyticsRepository
	codes       port.CodeGenerator
	clock       port.Clock
	metadata    port.MetadataFetcher
	classifier  port.RequestClassifier
	baseURL     string
	maxRetries  int
}

type NewServiceInput struct {
	ShortURLs   repository.ShortURLRepository
	Clicks      repository.ClickRepository
	Conversions repository.ConversionRepository
	Analytics   repository.AnalyticsRepository
	Codes       port.CodeGenerator
	Clock       port.Clock
	Metadata    port.MetadataFetcher
	Classifier  port.RequestClassifier
	BaseURL     string
	MaxRetries  int
}

func NewService(input NewServiceInput) *Service {
	retries := input.MaxRetries
	if retries <= 0 {
		retries = DefaultCollisionRetries
	}
	return &Service{
		shortURLs:   input.ShortURLs,
		clicks:      input.Clicks,
		conversions: input.Conversions,
		analytics:   input.Analytics,
		codes:       input.Codes,
		clock:       input.Clock,
		metadata:    input.Metadata,
		classifier:  input.Classifier,
		baseURL:     strings.TrimRight(input.BaseURL, "/"),
		maxRetries:  retries,
	}
}

type CreateShortURLInput struct {
	URL          string
	ReferralCode string
	Campaign     string
}

type CreateShortURLOutput struct {
	ShortURL string
	Entity   domain.ShortURL
}

func (s *Service) CreateShortURL(ctx context.Context, input CreateShortURLInput) (CreateShortURLOutput, error) {
	if !isHTTPURL(input.URL) {
		return CreateShortURLOutput{}, ErrInvalidURL
	}

	preview, err := s.metadata.Fetch(ctx, input.URL)
	if err != nil {
		preview = fallbackPreview(input.URL)
	} else {
		preview = mergePreview(preview, input.URL)
	}

	now := s.clock.Now()
	for attempt := 0; attempt < s.maxRetries; attempt++ {
		code, err := s.codes.Generate()
		if err != nil {
			return CreateShortURLOutput{}, err
		}

		entity := domain.ShortURL{
			Code:         code,
			OriginalURL:  input.URL,
			ReferralCode: strings.TrimSpace(input.ReferralCode),
			Campaign:     strings.TrimSpace(input.Campaign),
			Preview:      preview,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		created, err := s.shortURLs.Create(ctx, entity)
		if errors.Is(err, repository.ErrDuplicateCode) {
			continue
		}
		if err != nil {
			return CreateShortURLOutput{}, err
		}
		return CreateShortURLOutput{
			ShortURL: s.baseURL + "/" + created.Code,
			Entity:   created,
		}, nil
	}

	return CreateShortURLOutput{}, ErrCodeCollision
}

type ResolveInput struct {
	Code      string
	UserAgent string
	Referer   string
	Headers   map[string]string
}

type ResolveOutput struct {
	ShortURL  domain.ShortURL
	IsPreview bool
}

func (s *Service) Resolve(ctx context.Context, input ResolveInput) (ResolveOutput, error) {
	entity, ok, err := s.shortURLs.GetByCode(ctx, input.Code)
	if err != nil {
		return ResolveOutput{}, err
	}
	if !ok {
		return ResolveOutput{}, ErrShortURLNotFound
	}

	isPreview := s.classifier.IsCrawler(input.UserAgent)
	if isPreview {
		preview, err := s.metadata.Fetch(ctx, entity.OriginalURL)
		if err == nil {
			entity.Preview = mergePreview(preview, entity.OriginalURL)
			entity.UpdatedAt = s.clock.Now()
			_ = s.shortURLs.UpdatePreview(ctx, entity.ID, entity.Preview, entity.UpdatedAt)
		}
	} else {
		_ = s.clicks.Record(ctx, domain.ClickEvent{
			ShortURLID:   entity.ID,
			ReferralCode: entity.ReferralCode,
			Campaign:     entity.Campaign,
			UserAgent:    input.UserAgent,
			Platform:     s.classifier.Platform(input.UserAgent, input.Referer),
			Device:       s.classifier.Device(input.UserAgent),
			Country:      s.classifier.Country(input.Headers),
			OccurredAt:   s.clock.Now(),
		})
	}

	return ResolveOutput{ShortURL: entity, IsPreview: isPreview}, nil
}

type RecordConversionInput struct {
	Code         string
	ReferralCode string
	Campaign     string
	Value        float64
}

func (s *Service) RecordConversion(ctx context.Context, input RecordConversionInput) error {
	entity, ok, err := s.shortURLs.GetByCode(ctx, input.Code)
	if err != nil {
		return err
	}
	if !ok {
		return ErrShortURLNotFound
	}

	referral := strings.TrimSpace(input.ReferralCode)
	if referral == "" {
		referral = entity.ReferralCode
	}
	campaign := strings.TrimSpace(input.Campaign)
	if campaign == "" {
		campaign = entity.Campaign
	}

	return s.conversions.RecordConversion(ctx, domain.ConversionEvent{
		ShortURLID:   entity.ID,
		ReferralCode: referral,
		Campaign:     campaign,
		Value:        input.Value,
		OccurredAt:   s.clock.Now(),
	})
}

func (s *Service) Analytics(ctx context.Context, code string) (domain.AnalyticsSummary, error) {
	summary, ok, err := s.analytics.Summary(ctx, code)
	if err != nil {
		return domain.AnalyticsSummary{}, err
	}
	if !ok {
		return domain.AnalyticsSummary{}, ErrShortURLNotFound
	}
	return summary, nil
}

func (s *Service) ShortURLForCode(code string) string {
	return fmt.Sprintf("%s/%s", s.baseURL, code)
}

func isHTTPURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func fallbackPreview(rawURL string) domain.PreviewMetadata {
	return domain.PreviewMetadata{
		Title:       "Shared link",
		Description: rawURL,
	}
}

func mergePreview(preview domain.PreviewMetadata, rawURL string) domain.PreviewMetadata {
	if strings.TrimSpace(preview.Title) == "" {
		preview.Title = "Shared link"
	}
	if strings.TrimSpace(preview.Description) == "" {
		preview.Description = rawURL
	}
	return preview
}
