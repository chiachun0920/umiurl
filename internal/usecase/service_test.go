package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"umiurl/internal/domain"
	"umiurl/internal/usecase/interface/repository"
)

func TestCreateShortURLRetriesDuplicateCode(t *testing.T) {
	repo := newMemoryRepo()
	generator := &sequenceGenerator{codes: []string{"AAAAAAA", "BBBBBBB"}}
	metadata := &staticMetadata{}
	service := testService(repo, generator, metadata)

	repo.duplicateCodes["AAAAAAA"] = true

	output, err := service.CreateShortURL(context.Background(), CreateShortURLInput{
		URL:          "https://example.com/post",
		ReferralCode: "user_001",
		Campaign:     "spring",
	})
	if err != nil {
		t.Fatalf("CreateShortURL() error = %v", err)
	}
	if output.Entity.Code != "BBBBBBB" {
		t.Fatalf("code = %q, want BBBBBBB", output.Entity.Code)
	}
	if output.ShortURL != "http://short.test/BBBBBBB" {
		t.Fatalf("short url = %q", output.ShortURL)
	}
	if metadata.calls != 1 {
		t.Fatalf("metadata calls = %d, want 1", metadata.calls)
	}
	if output.Entity.Preview.Title != "Example" {
		t.Fatalf("preview title = %q", output.Entity.Preview.Title)
	}
}

func TestCreateShortURLFallsBackWhenMetadataFetchFails(t *testing.T) {
	repo := newMemoryRepo()
	metadata := &staticMetadata{err: errors.New("fetch failed")}
	service := testService(repo, &sequenceGenerator{codes: []string{"AAAAAAA"}}, metadata)

	output, err := service.CreateShortURL(context.Background(), CreateShortURLInput{URL: "https://example.com/post"})
	if err != nil {
		t.Fatalf("CreateShortURL() error = %v", err)
	}
	if output.Entity.Preview.Title != "Shared link" {
		t.Fatalf("preview title = %q", output.Entity.Preview.Title)
	}
	if output.Entity.Preview.Description != "https://example.com/post" {
		t.Fatalf("preview description = %q", output.Entity.Preview.Description)
	}
}

func TestCreateShortURLRejectsInvalidURL(t *testing.T) {
	metadata := &staticMetadata{}
	service := testService(newMemoryRepo(), &sequenceGenerator{codes: []string{"AAAAAAA"}}, metadata)

	_, err := service.CreateShortURL(context.Background(), CreateShortURLInput{URL: "ftp://example.com"})
	if !errors.Is(err, ErrInvalidURL) {
		t.Fatalf("error = %v, want ErrInvalidURL", err)
	}
	if metadata.calls != 0 {
		t.Fatalf("metadata calls = %d, want 0", metadata.calls)
	}
}

func TestResolveRecordsClickForNonCrawler(t *testing.T) {
	repo := newMemoryRepo()
	repo.urls["AAAAAAA"] = domain.ShortURL{
		ID:           1,
		Code:         "AAAAAAA",
		OriginalURL:  "https://example.com",
		ReferralCode: "user_001",
		Campaign:     "spring",
	}
	metadata := &staticMetadata{}
	service := testService(repo, &sequenceGenerator{codes: []string{"BBBBBBB"}}, metadata)

	output, err := service.Resolve(context.Background(), ResolveInput{
		Code:      "AAAAAAA",
		UserAgent: "Mozilla/5.0 iPhone",
		Referer:   "https://facebook.com",
		Headers:   map[string]string{"X-Country": "tw"},
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if output.IsPreview {
		t.Fatal("IsPreview = true, want false")
	}
	if len(repo.clicks) != 1 {
		t.Fatalf("click count = %d, want 1", len(repo.clicks))
	}
	if metadata.calls != 0 {
		t.Fatalf("metadata calls = %d, want 0", metadata.calls)
	}
	click := repo.clicks[0]
	if click.Platform != "facebook" || click.Device != "mobile" || click.Country != "TW" {
		t.Fatalf("click classification = %#v", click)
	}
}

func TestResolveDoesNotRecordClickForCrawler(t *testing.T) {
	repo := newMemoryRepo()
	repo.urls["AAAAAAA"] = domain.ShortURL{ID: 1, Code: "AAAAAAA", OriginalURL: "https://example.com"}
	metadata := &staticMetadata{preview: domain.PreviewMetadata{Title: "Fetched title", Description: "Fetched description"}}
	service := testService(repo, &sequenceGenerator{codes: []string{"BBBBBBB"}}, metadata)

	output, err := service.Resolve(context.Background(), ResolveInput{
		Code:      "AAAAAAA",
		UserAgent: "facebookexternalhit/1.1",
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !output.IsPreview {
		t.Fatal("IsPreview = false, want true")
	}
	if len(repo.clicks) != 0 {
		t.Fatalf("click count = %d, want 0", len(repo.clicks))
	}
	if metadata.calls != 1 {
		t.Fatalf("metadata calls = %d, want 1", metadata.calls)
	}
	if output.ShortURL.Preview.Title != "Fetched title" {
		t.Fatalf("preview title = %q", output.ShortURL.Preview.Title)
	}
	if repo.previewUpdates != 1 {
		t.Fatalf("preview updates = %d, want 1", repo.previewUpdates)
	}
}

func TestResolveCrawlerFallsBackWhenMetadataFetchFails(t *testing.T) {
	repo := newMemoryRepo()
	repo.urls["AAAAAAA"] = domain.ShortURL{
		ID:          1,
		Code:        "AAAAAAA",
		OriginalURL: "https://example.com",
		Preview:     domain.PreviewMetadata{Title: "Stored title", Description: "Stored description"},
	}
	metadata := &staticMetadata{err: errors.New("fetch failed")}
	service := testService(repo, &sequenceGenerator{codes: []string{"BBBBBBB"}}, metadata)

	output, err := service.Resolve(context.Background(), ResolveInput{
		Code:      "AAAAAAA",
		UserAgent: "facebookexternalhit/1.1",
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if output.ShortURL.Preview.Title != "Stored title" {
		t.Fatalf("preview title = %q", output.ShortURL.Preview.Title)
	}
	if repo.previewUpdates != 0 {
		t.Fatalf("preview updates = %d, want 0", repo.previewUpdates)
	}
}

func testService(repo *memoryRepo, generator *sequenceGenerator, metadata *staticMetadata) *Service {
	return NewService(NewServiceInput{
		ShortURLs:   repo,
		Clicks:      repo,
		Conversions: repo,
		Analytics:   repo,
		Codes:       generator,
		Clock:       fixedClock{},
		Metadata:    metadata,
		Classifier:  testClassifier{},
		BaseURL:     "http://short.test",
		MaxRetries:  5,
	})
}

type memoryRepo struct {
	urls           map[string]domain.ShortURL
	duplicateCodes map[string]bool
	clicks         []domain.ClickEvent
	conversions    []domain.ConversionEvent
	previewUpdates int
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		urls:           map[string]domain.ShortURL{},
		duplicateCodes: map[string]bool{},
	}
}

func (r *memoryRepo) Create(_ context.Context, shortURL domain.ShortURL) (domain.ShortURL, error) {
	if r.duplicateCodes[shortURL.Code] {
		return domain.ShortURL{}, repository.ErrDuplicateCode
	}
	shortURL.ID = int64(len(r.urls) + 1)
	r.urls[shortURL.Code] = shortURL
	return shortURL, nil
}

func (r *memoryRepo) GetByCode(_ context.Context, code string) (domain.ShortURL, bool, error) {
	entity, ok := r.urls[code]
	return entity, ok, nil
}

func (r *memoryRepo) UpdatePreview(_ context.Context, id int64, preview domain.PreviewMetadata, updatedAt time.Time) error {
	for code, entity := range r.urls {
		if entity.ID == id {
			entity.Preview = preview
			entity.UpdatedAt = updatedAt
			r.urls[code] = entity
			r.previewUpdates++
			return nil
		}
	}
	return nil
}

func (r *memoryRepo) Record(_ context.Context, event domain.ClickEvent) error {
	r.clicks = append(r.clicks, event)
	return nil
}

func (r *memoryRepo) RecordConversion(_ context.Context, event domain.ConversionEvent) error {
	r.conversions = append(r.conversions, event)
	return nil
}

func (r *memoryRepo) Summary(_ context.Context, code string) (domain.AnalyticsSummary, bool, error) {
	_, ok := r.urls[code]
	return domain.AnalyticsSummary{Code: code}, ok, nil
}

type sequenceGenerator struct {
	codes []string
	next  int
}

func (g *sequenceGenerator) Generate() (string, error) {
	code := g.codes[g.next]
	g.next++
	return code, nil
}

type fixedClock struct{}

func (fixedClock) Now() time.Time {
	return time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
}

type staticMetadata struct {
	preview domain.PreviewMetadata
	err     error
	calls   int
}

func (m *staticMetadata) Fetch(context.Context, string) (domain.PreviewMetadata, error) {
	m.calls++
	if m.err != nil {
		return domain.PreviewMetadata{}, m.err
	}
	if m.preview.Title != "" || m.preview.Description != "" || m.preview.ImageURL != "" {
		return m.preview, nil
	}
	return domain.PreviewMetadata{Title: "Example", Description: "Description"}, nil
}

type testClassifier struct{}

func (testClassifier) IsCrawler(userAgent string) bool {
	return userAgent == "facebookexternalhit/1.1"
}

func (testClassifier) Platform(_, referer string) string {
	if referer == "https://facebook.com" {
		return "facebook"
	}
	return "direct"
}

func (testClassifier) Device(userAgent string) string {
	if userAgent == "Mozilla/5.0 iPhone" {
		return "mobile"
	}
	return "desktop"
}

func (testClassifier) Country(headers map[string]string) string {
	return "TW"
}
