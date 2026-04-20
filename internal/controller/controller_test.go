package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"umiurl/internal/domain"
	"umiurl/internal/usecase"
	"umiurl/internal/usecase/interface/repository"
)

func TestCreateShortURL(t *testing.T) {
	e, repo := testEcho()

	body := bytes.NewBufferString(`{"url":"https://example.com","referral_code":"user_001","campaign":"spring"}`)
	req := httptest.NewRequest(http.MethodPost, "/urls", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var response createURLResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Code != "AAAAAAA" || response.ShortURL != "http://short.test/AAAAAAA" {
		t.Fatalf("response = %#v", response)
	}
	if len(repo.urls) != 1 {
		t.Fatalf("url count = %d, want 1", len(repo.urls))
	}
}

func TestResolveRedirect(t *testing.T) {
	e, repo := testEcho()
	repo.urls["AAAAAAA"] = domain.ShortURL{ID: 1, Code: "AAAAAAA", OriginalURL: "https://example.com"}

	req := httptest.NewRequest(http.MethodGet, "/AAAAAAA", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if location := rec.Header().Get("Location"); location != "https://example.com" {
		t.Fatalf("Location = %q", location)
	}
}

func TestResolvePreview(t *testing.T) {
	e, repo := testEcho()
	repo.urls["AAAAAAA"] = domain.ShortURL{
		ID:          1,
		Code:        "AAAAAAA",
		OriginalURL: "https://example.com",
		Preview:     domain.PreviewMetadata{Title: "Preview title", Description: "Preview description"},
	}

	req := httptest.NewRequest(http.MethodGet, "/AAAAAAA", nil)
	req.Header.Set("User-Agent", "facebookexternalhit/1.1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`og:title`)) {
		t.Fatalf("preview html missing og:title: %s", rec.Body.String())
	}
}

func testEcho() (*echo.Echo, *controllerMemoryRepo) {
	repo := &controllerMemoryRepo{urls: map[string]domain.ShortURL{}}
	service := usecase.NewService(usecase.NewServiceInput{
		ShortURLs:   repo,
		Clicks:      repo,
		Conversions: repo,
		Analytics:   repo,
		Codes:       &controllerCodeGenerator{},
		Clock:       controllerClock{},
		Metadata:    controllerMetadata{},
		Classifier:  controllerClassifier{},
		BaseURL:     "http://short.test",
		MaxRetries:  5,
	})
	e := echo.New()
	New(service).Register(e)
	return e, repo
}

type controllerMemoryRepo struct {
	urls   map[string]domain.ShortURL
	clicks []domain.ClickEvent
}

func (r *controllerMemoryRepo) Create(_ context.Context, shortURL domain.ShortURL) (domain.ShortURL, error) {
	if _, ok := r.urls[shortURL.Code]; ok {
		return domain.ShortURL{}, repository.ErrDuplicateCode
	}
	shortURL.ID = int64(len(r.urls) + 1)
	r.urls[shortURL.Code] = shortURL
	return shortURL, nil
}

func (r *controllerMemoryRepo) GetByCode(_ context.Context, code string) (domain.ShortURL, bool, error) {
	entity, ok := r.urls[code]
	return entity, ok, nil
}

func (r *controllerMemoryRepo) UpdatePreview(_ context.Context, id int64, preview domain.PreviewMetadata, updatedAt time.Time) error {
	for code, entity := range r.urls {
		if entity.ID == id {
			entity.Preview = preview
			entity.UpdatedAt = updatedAt
			r.urls[code] = entity
			return nil
		}
	}
	return nil
}

func (r *controllerMemoryRepo) Record(_ context.Context, event domain.ClickEvent) error {
	r.clicks = append(r.clicks, event)
	return nil
}

func (r *controllerMemoryRepo) RecordConversion(context.Context, domain.ConversionEvent) error {
	return nil
}

func (r *controllerMemoryRepo) Summary(_ context.Context, code string) (domain.AnalyticsSummary, bool, error) {
	_, ok := r.urls[code]
	return domain.AnalyticsSummary{Code: code}, ok, nil
}

type controllerCodeGenerator struct{}

func (*controllerCodeGenerator) Generate() (string, error) {
	return "AAAAAAA", nil
}

type controllerClock struct{}

func (controllerClock) Now() time.Time {
	return time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
}

type controllerMetadata struct{}

func (controllerMetadata) Fetch(context.Context, string) (domain.PreviewMetadata, error) {
	return domain.PreviewMetadata{Title: "Preview", Description: "Description"}, nil
}

type controllerClassifier struct{}

func (controllerClassifier) IsCrawler(userAgent string) bool {
	return userAgent == "facebookexternalhit/1.1"
}

func (controllerClassifier) Platform(string, string) string {
	return "direct"
}

func (controllerClassifier) Device(string) string {
	return "desktop"
}

func (controllerClassifier) Country(map[string]string) string {
	return "unknown"
}
