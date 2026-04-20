package gateway

import "strings"

type SimpleRequestClassifier struct{}

func (SimpleRequestClassifier) IsCrawler(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	crawlers := []string{
		"facebookexternalhit",
		"twitterbot",
		"linkedinbot",
		"slackbot",
		"discordbot",
		"telegrambot",
		"line/",
		"whatsapp",
		"bot",
		"crawler",
		"spider",
	}
	for _, crawler := range crawlers {
		if strings.Contains(ua, crawler) {
			return true
		}
	}
	return false
}

func (SimpleRequestClassifier) Platform(userAgent, referer string) string {
	ref := strings.ToLower(referer)
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ref, "facebook") || strings.Contains(ua, "facebook"):
		return "facebook"
	case strings.Contains(ref, "threads") || strings.Contains(ua, "threads"):
		return "threads"
	case strings.Contains(ref, "twitter") || strings.Contains(ref, "x.com") || strings.Contains(ua, "twitter"):
		return "x"
	case strings.Contains(ref, "linkedin") || strings.Contains(ua, "linkedin"):
		return "linkedin"
	default:
		return "direct"
	}
}

func (SimpleRequestClassifier) Device(userAgent string) string {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "mobile") || strings.Contains(ua, "iphone") || strings.Contains(ua, "android"):
		return "mobile"
	case strings.Contains(ua, "ipad") || strings.Contains(ua, "tablet"):
		return "tablet"
	default:
		return "desktop"
	}
}

func (SimpleRequestClassifier) Country(headers map[string]string) string {
	for _, key := range []string{"CF-IPCountry", "X-Country", "CloudFront-Viewer-Country"} {
		if value := strings.TrimSpace(headers[key]); value != "" {
			return strings.ToUpper(value)
		}
	}
	return "unknown"
}
