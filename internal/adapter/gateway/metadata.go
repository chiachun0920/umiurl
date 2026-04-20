package gateway

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"umiurl/internal/domain"
)

type HTTPMetadataFetcher struct {
	client *http.Client
}

const metadataMaxBytes = 2 * 1024 * 1024

func NewHTTPMetadataFetcher() *HTTPMetadataFetcher {
	return &HTTPMetadataFetcher{
		client: &http.Client{Timeout: 3 * time.Second},
	}
}

func (f *HTTPMetadataFetcher) Fetch(ctx context.Context, rawURL string) (domain.PreviewMetadata, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return domain.PreviewMetadata{}, err
	}
	req.Header.Set("User-Agent", "umiurl-preview-bot/1.0")

	fmt.Println("set header")

	resp, err := f.client.Do(req)
	if err != nil {
		return domain.PreviewMetadata{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return domain.PreviewMetadata{}, fmt.Errorf("fetch metadata status %d", resp.StatusCode)
	}

	return parsePreviewMetadata(rawURL, io.LimitReader(resp.Body, metadataMaxBytes))
}

func parsePreviewMetadata(rawURL string, reader io.Reader) (domain.PreviewMetadata, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return domain.PreviewMetadata{}, err
	}

	ogTags := map[string]string{}
	doc.Find("meta").Each(func(_ int, selection *goquery.Selection) {
		property, _ := selection.Attr("property")
		property = strings.ToLower(strings.TrimSpace(property))
		if !strings.HasPrefix(property, "og:") {
			return
		}
		content, _ := selection.Attr("content")
		ogTags[property] = strings.TrimSpace(content)
	})

	return domain.PreviewMetadata{
		Title:       firstNonEmpty(ogTags["og:title"], doc.Find("title").First().Text(), "Shared link"),
		Description: firstNonEmpty(ogTags["og:description"], metaName(doc, "description"), rawURL),
		ImageURL:    ogTags["og:image"],
	}, nil
}

func metaName(doc *goquery.Document, name string) string {
	var result string
	doc.Find("meta").EachWithBreak(func(_ int, selection *goquery.Selection) bool {
		attrName, _ := selection.Attr("name")
		if !strings.EqualFold(strings.TrimSpace(attrName), name) {
			return true
		}
		content, _ := selection.Attr("content")
		result = strings.TrimSpace(content)
		return false
	})
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
