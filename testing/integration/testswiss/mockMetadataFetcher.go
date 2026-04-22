package testswiss

import (
	"context"

	"umiurl/internal/domain"
)

type MockMetadataFetcher struct{}

func NewMockMetadataFetcher() *MockMetadataFetcher {
	return &MockMetadataFetcher{}
}

func (f *MockMetadataFetcher) Fetch(ctx context.Context, url string) (domain.PreviewMetadata, error) {
	return domain.PreviewMetadata{
		Title:       "Example Domain",
		Description: "This domain is for use in illustrative examples in documents.",
		ImageURL:    "https://www.example.com/image.png",
	}, nil
}
