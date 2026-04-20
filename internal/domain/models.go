package domain

import "time"

type PreviewMetadata struct {
	Title       string
	Description string
	ImageURL    string
}

type ShortURL struct {
	ID           int64
	Code         string
	OriginalURL  string
	ReferralCode string
	Campaign     string
	Preview      PreviewMetadata
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ClickEvent struct {
	ShortURLID   int64
	ReferralCode string
	Campaign     string
	UserAgent    string
	Platform     string
	Device       string
	Country      string
	OccurredAt   time.Time
}

type ConversionEvent struct {
	ShortURLID   int64
	ReferralCode string
	Campaign     string
	Value        float64
	OccurredAt   time.Time
}

type Breakdown struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

type AnalyticsSummary struct {
	Code              string      `json:"code"`
	TotalClicks       int64       `json:"total_clicks"`
	TotalConversions  int64       `json:"total_conversions"`
	ClicksByReferral  []Breakdown `json:"clicks_by_referral"`
	ClicksByCampaign  []Breakdown `json:"clicks_by_campaign"`
	ClicksByPlatform  []Breakdown `json:"clicks_by_platform"`
	ClicksByDevice    []Breakdown `json:"clicks_by_device"`
	ClicksByCountry   []Breakdown `json:"clicks_by_country"`
	ConversionsByRef  []Breakdown `json:"conversions_by_referral"`
	ConversionsByCamp []Breakdown `json:"conversions_by_campaign"`
}
