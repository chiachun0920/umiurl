package shorturl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"umiurl/internal/controller"
	"umiurl/internal/domain"

	"github.com/labstack/echo/v4"
)

func (shortUrlSfn *ShortUrlStepFunc) iCreateAShortUrlWith(
	ctx context.Context,
	longUrl string,
) (context.Context, error) {
	req := controller.CreateURLRequest{
		URL:          longUrl,
		ReferralCode: "user_001",
		Campaign:     "spring",
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return ctx, fmt.Errorf("marshal request: %v", err)
	}
	body := bytes.NewBuffer(reqBody)
	httpreq := httptest.NewRequest(http.MethodPost, "/urls", body)
	httpreq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	shortUrlSfn.Swiss.EchoServer.ServeHTTP(rec, httpreq)

	ctx = context.WithValue(ctx, statusCodeCtxKey, rec.Code)
	ctx = context.WithValue(ctx, responseCtxKey, rec.Body.String())

	return ctx, nil
}

func (shortUrlSfn *ShortUrlStepFunc) iShouldReceiveAShortUrl(
	ctx context.Context,
) (context.Context, error) {
	if ctx.Value(statusCodeCtxKey) != http.StatusCreated {
		return ctx, fmt.Errorf("status = %d, body = %s", ctx.Value(statusCodeCtxKey), ctx.Value(responseCtxKey))
	}

	res := controller.CreateURLResponse{}
	if err := json.Unmarshal([]byte(ctx.Value(responseCtxKey).(string)), &res); err != nil {
		return ctx, fmt.Errorf("decode response: %v", err)
	}

	if res.ShortURL == "" {
		return ctx, fmt.Errorf("short url is empty")
	}

	return ctx, nil
}

func (shortUrlSfn *ShortUrlStepFunc) theShortUrlShouldRedirectTo(
	ctx context.Context,
	expectedURL string,
) (context.Context, error) {
	res := controller.CreateURLResponse{}
	if err := json.Unmarshal([]byte(ctx.Value(responseCtxKey).(string)), &res); err != nil {
		return ctx, fmt.Errorf("decode response: %v", err)
	}

	httpreq := httptest.NewRequest(http.MethodGet, res.ShortURL, nil)
	rec := httptest.NewRecorder()
	shortUrlSfn.Swiss.EchoServer.ServeHTTP(rec, httpreq)

	if rec.Code != http.StatusFound {
		return ctx, fmt.Errorf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	location := rec.Header().Get(echo.HeaderLocation)
	if location != expectedURL {
		return ctx, fmt.Errorf("location = %s, want %s", location, expectedURL)
	}

	return ctx, nil
}

func (shortUrlSfn *ShortUrlStepFunc) iShouldReceiveAnErrorMessage(
	ctx context.Context,
	expectedMessage string,
) (context.Context, error) {
	if ctx.Value(statusCodeCtxKey) == http.StatusCreated {
		return ctx, fmt.Errorf("expected error but got status %d", ctx.Value(statusCodeCtxKey))
	}

	var errResp map[string]string
	if err := json.Unmarshal([]byte(ctx.Value(responseCtxKey).(string)), &errResp); err != nil {
		return ctx, fmt.Errorf("decode error response: %v", err)
	}

	fmt.Println(errResp)
	if errResp["error"] != expectedMessage {
		return ctx, fmt.Errorf("error message = %s, want %s", errResp["error"], expectedMessage)
	}

	return ctx, nil
}

func (shortUrlSfn *ShortUrlStepFunc) aShortUrlExistsForWithCode(
	ctx context.Context,
	originalURL string,
	code string,
) (context.Context, error) {
	shortUrlSfn.Swiss.Repo.Create(ctx, domain.ShortURL{
		Code:        code,
		OriginalURL: originalURL,
	})

	return ctx, nil
}

func (shortUrlSfn *ShortUrlStepFunc) iViewTheAnalyticsForTheShortUrlWithCode(
	ctx context.Context,
	code string,
) (context.Context, error) {
	httpreq := httptest.NewRequest(http.MethodGet, "/urls/"+code+"/analytics", nil)
	rec := httptest.NewRecorder()
	shortUrlSfn.Swiss.EchoServer.ServeHTTP(rec, httpreq)

	ctx = context.WithValue(ctx, statusCodeCtxKey, rec.Code)
	ctx = context.WithValue(ctx, responseCtxKey, rec.Body.String())

	return ctx, nil
}

func (shortUrlSfn *ShortUrlStepFunc) iShouldReceiveAnalyticsDataForTheShortUrlWithCode(
	ctx context.Context,
	code string,
) (context.Context, error) {
	if ctx.Value(statusCodeCtxKey) != http.StatusOK {
		return ctx, fmt.Errorf("status = %d, body = %s", ctx.Value(statusCodeCtxKey), ctx.Value(responseCtxKey))
	}

	fmt.Println("-->", ctx.Value(responseCtxKey).(string))
	var summary domain.AnalyticsSummary
	if err := json.Unmarshal([]byte(ctx.Value(responseCtxKey).(string)), &summary); err != nil {
		return ctx, fmt.Errorf("decode response: %v", err)
	}

	if summary.Code != code {
		return ctx, fmt.Errorf("analytics code = %s, want %s", summary.Code, code)
	}

	return ctx, nil
}

func (shortUrlSfn *ShortUrlStepFunc) iVisitTheShortUrlWithCode(
	ctx context.Context,
	code string,
) (context.Context, error) {
	httpreq := httptest.NewRequest(http.MethodGet, "/"+code, nil)
	rec := httptest.NewRecorder()
	shortUrlSfn.Swiss.EchoServer.ServeHTTP(rec, httpreq)

	if rec.Code != http.StatusFound {
		return ctx, fmt.Errorf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	return ctx, nil
}

func (shortUrlSfn *ShortUrlStepFunc) iShouldReceiveAnalyticsDataForTheShortUrlWithCodeWithVisit(
	ctx context.Context,
	code string,
	expectedVisit int,
) (context.Context, error) {
	if ctx.Value(statusCodeCtxKey) != http.StatusOK {
		return ctx, fmt.Errorf("status = %d, body = %s", ctx.Value(statusCodeCtxKey), ctx.Value(responseCtxKey))
	}

	var summary domain.AnalyticsSummary
	if err := json.Unmarshal([]byte(ctx.Value(responseCtxKey).(string)), &summary); err != nil {
		return ctx, fmt.Errorf("decode response: %v", err)
	}

	if summary.Code != code {
		return ctx, fmt.Errorf("analytics code = %s, want %s", summary.Code, code)
	}
	if summary.TotalClicks != int64(expectedVisit) {
		return ctx, fmt.Errorf("total clicks = %d, want %d", summary.TotalClicks, expectedVisit)
	}

	return ctx, nil
}
