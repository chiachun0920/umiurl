package shorturl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"umiurl/internal/controller"

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
