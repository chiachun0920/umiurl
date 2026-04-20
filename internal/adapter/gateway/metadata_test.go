package gateway

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestHTTPMetadataFetcherFetchesOGTagsWithGoquery(t *testing.T) {
	preview, err := parsePreviewMetadata("https://example.com", strings.NewReader(`<!doctype html>
<html>
<head>
  <meta property="og:title" content="OG Title">
  <meta property="og:description" content="OG Description">
  <meta property="og:image" content="https://example.com/og.png">
</head>
</html>`))
	if err != nil {
		t.Fatalf("parsePreviewMetadata() error = %v", err)
	}
	if preview.Title != "OG Title" {
		t.Fatalf("Title = %q", preview.Title)
	}
	if preview.Description != "OG Description" {
		t.Fatalf("Description = %q", preview.Description)
	}
	if preview.ImageURL != "https://example.com/og.png" {
		t.Fatalf("ImageURL = %q", preview.ImageURL)
	}
}

func TestHTTPMetadataFetcherFallsBackToStandardHTMLTags(t *testing.T) {
	preview, err := parsePreviewMetadata("https://example.com", strings.NewReader(`<!doctype html>
<html>
<head>
  <title>HTML Title</title>
  <meta name="description" content="HTML Description">
</head>
</html>`))
	if err != nil {
		t.Fatalf("parsePreviewMetadata() error = %v", err)
	}
	if preview.Title != "HTML Title" {
		t.Fatalf("Title = %q", preview.Title)
	}
	if preview.Description != "HTML Description" {
		t.Fatalf("Description = %q", preview.Description)
	}
}

func TestHTTPMetadataFetcherFallsBackWhenMetadataMissing(t *testing.T) {
	rawURL := "https://example.com"
	preview, err := parsePreviewMetadata(rawURL, strings.NewReader(`<!doctype html><html><head></head><body></body></html>`))
	if err != nil {
		t.Fatalf("parsePreviewMetadata() error = %v", err)
	}
	if preview.Title != "Shared link" {
		t.Fatalf("Title = %q", preview.Title)
	}
	if preview.Description != rawURL {
		t.Fatalf("Description = %q", preview.Description)
	}
}

func TestHTTPMetadataFetcherReturnsErrorOnNon2xx(t *testing.T) {
	fetcher := &HTTPMetadataFetcher{
		client: &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewBufferString("not found")),
				}, nil
			}),
		},
	}
	_, err := fetcher.Fetch(context.Background(), "https://example.com")
	if err == nil {
		t.Fatal("Fetch() error = nil, want error")
	}
}

func TestHTTPMetadataFetcherReadsYouTubeSizedMetadata(t *testing.T) {
	padding := strings.Repeat("x", 600*1024)
	fetcher := &HTTPMetadataFetcher{
		client: &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`<!doctype html>
<html>
<head>` + padding + `
  <meta property="og:title" content="Late OG Title">
  <meta property="og:description" content="Late OG Description">
  <meta property="og:image" content="https://example.com/late.png">
</head>
</html>`)),
				}, nil
			}),
		},
	}

	preview, err := fetcher.Fetch(context.Background(), "https://www.youtube.com/watch?v=BkszA-MvjXA")
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if preview.Title != "Late OG Title" {
		t.Fatalf("Title = %q", preview.Title)
	}
	if preview.Description != "Late OG Description" {
		t.Fatalf("Description = %q", preview.Description)
	}
	if preview.ImageURL != "https://example.com/late.png" {
		t.Fatalf("ImageURL = %q", preview.ImageURL)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
