package controller

import (
	"errors"
	"html"
	"net/http"

	"github.com/labstack/echo/v4"

	"umiurl/internal/domain"
	"umiurl/internal/usecase"
)

type Controller struct {
	service *usecase.Service
}

func New(service *usecase.Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) Register(e *echo.Echo) {
	e.POST("/urls", c.createShortURL)
	e.GET("/urls/:code/analytics", c.analytics)
	e.POST("/conversions", c.recordConversion)
	e.GET("/:code", c.resolve)
}

type createURLRequest struct {
	URL          string `json:"url"`
	ReferralCode string `json:"referral_code"`
	Campaign     string `json:"campaign"`
}

type createURLResponse struct {
	Code         string          `json:"code"`
	ShortURL     string          `json:"short_url"`
	URL          string          `json:"url"`
	ReferralCode string          `json:"referral_code,omitempty"`
	Campaign     string          `json:"campaign,omitempty"`
	Preview      previewResponse `json:"preview"`
}

type previewResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url,omitempty"`
}

func (c *Controller) createShortURL(ctx echo.Context) error {
	var req createURLRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse("invalid json body"))
	}

	output, err := c.service.CreateShortURL(ctx.Request().Context(), usecase.CreateShortURLInput{
		URL:          req.URL,
		ReferralCode: req.ReferralCode,
		Campaign:     req.Campaign,
	})
	if errors.Is(err, usecase.ErrInvalidURL) {
		return ctx.JSON(http.StatusBadRequest, errorResponse("url must be absolute http or https"))
	}
	if errors.Is(err, usecase.ErrCodeCollision) {
		return ctx.JSON(http.StatusInternalServerError, errorResponse("could not allocate short code"))
	}
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, errorResponse("create short url failed"))
	}

	entity := output.Entity
	return ctx.JSON(http.StatusCreated, createURLResponse{
		Code:         entity.Code,
		ShortURL:     output.ShortURL,
		URL:          entity.OriginalURL,
		ReferralCode: entity.ReferralCode,
		Campaign:     entity.Campaign,
		Preview:      toPreviewResponse(entity.Preview),
	})
}

func (c *Controller) resolve(ctx echo.Context) error {
	headers := map[string]string{
		"CF-IPCountry":              ctx.Request().Header.Get("CF-IPCountry"),
		"X-Country":                 ctx.Request().Header.Get("X-Country"),
		"CloudFront-Viewer-Country": ctx.Request().Header.Get("CloudFront-Viewer-Country"),
	}

	output, err := c.service.Resolve(ctx.Request().Context(), usecase.ResolveInput{
		Code:      ctx.Param("code"),
		UserAgent: ctx.Request().UserAgent(),
		Referer:   ctx.Request().Referer(),
		Headers:   headers,
	})
	if errors.Is(err, usecase.ErrShortURLNotFound) {
		return ctx.JSON(http.StatusNotFound, errorResponse("short url not found"))
	}
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, errorResponse("resolve short url failed"))
	}

	if output.IsPreview {
		return ctx.HTML(http.StatusOK, previewHTML(output.ShortURL, c.service.ShortURLForCode(output.ShortURL.Code)))
	}
	return ctx.Redirect(http.StatusFound, output.ShortURL.OriginalURL)
}

func (c *Controller) analytics(ctx echo.Context) error {
	summary, err := c.service.Analytics(ctx.Request().Context(), ctx.Param("code"))
	if errors.Is(err, usecase.ErrShortURLNotFound) {
		return ctx.JSON(http.StatusNotFound, errorResponse("short url not found"))
	}
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, errorResponse("load analytics failed"))
	}
	return ctx.JSON(http.StatusOK, summary)
}

type conversionRequest struct {
	Code         string  `json:"code"`
	ReferralCode string  `json:"referral_code"`
	Campaign     string  `json:"campaign"`
	Value        float64 `json:"value"`
}

func (c *Controller) recordConversion(ctx echo.Context) error {
	var req conversionRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse("invalid json body"))
	}
	if req.Code == "" {
		return ctx.JSON(http.StatusBadRequest, errorResponse("code is required"))
	}

	err := c.service.RecordConversion(ctx.Request().Context(), usecase.RecordConversionInput{
		Code:         req.Code,
		ReferralCode: req.ReferralCode,
		Campaign:     req.Campaign,
		Value:        req.Value,
	})
	if errors.Is(err, usecase.ErrShortURLNotFound) {
		return ctx.JSON(http.StatusNotFound, errorResponse("short url not found"))
	}
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, errorResponse("record conversion failed"))
	}
	return ctx.JSON(http.StatusCreated, map[string]string{"status": "created"})
}

func toPreviewResponse(preview domain.PreviewMetadata) previewResponse {
	return previewResponse{
		Title:       preview.Title,
		Description: preview.Description,
		ImageURL:    preview.ImageURL,
	}
}

func previewHTML(entity domain.ShortURL, shortURL string) string {
	title := html.EscapeString(entity.Preview.Title)
	if title == "" {
		title = "Shared link"
	}
	description := html.EscapeString(entity.Preview.Description)
	imageURL := html.EscapeString(entity.Preview.ImageURL)
	canonical := html.EscapeString(shortURL)

	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>` + title + `</title>
  <meta property="og:title" content="` + title + `">
  <meta property="og:description" content="` + description + `">
  <meta property="og:url" content="` + canonical + `">
  <meta property="og:type" content="website">
  <meta name="twitter:card" content="summary_large_image">
  <meta name="twitter:title" content="` + title + `">
  <meta name="twitter:description" content="` + description + `">
  <meta property="og:image" content="` + imageURL + `">
  <meta name="twitter:image" content="` + imageURL + `">
</head>
<body></body>
</html>`
}

func errorResponse(message string) map[string]string {
	return map[string]string{"error": message}
}
