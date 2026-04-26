package shorturl

import (
	"umiurl/testing/integration/testswiss"

	"github.com/cucumber/godog"
)

type ContextKey string

const statusCodeCtxKey = ContextKey("statusCodeCtxKey")
const responseCtxKey = ContextKey("responseCtxKey")
const usersCtxKey = ContextKey("usersCtxKey")
const profileCtxKey = ContextKey("profileCtxKey")

type ShortUrlStepFunc struct {
	Swiss *testswiss.TestSwiss
}

func newShortUrlStepFunc(swiss *testswiss.TestSwiss) *ShortUrlStepFunc {
	return &ShortUrlStepFunc{Swiss: swiss}
}

func InitializeShortUrlScenario(ctx *godog.ScenarioContext, swiss *testswiss.TestSwiss) {
	sfh := newShortUrlStepFunc(swiss)

	ctx.Step(`^I create a short url with "([^"]*)"$`, sfh.iCreateAShortUrlWith)
	ctx.Step(`^I should receive a short url$`, sfh.iShouldReceiveAShortUrl)
	ctx.Step(`^the short url should redirect to "([^"]*)"$`, sfh.theShortUrlShouldRedirectTo)
	ctx.Step(`^I should receive an error message "([^"]*)"$`, sfh.iShouldReceiveAnErrorMessage)
	ctx.Step(`^a short url exists for "([^"]*)" with code "([^"]*)"$`, sfh.aShortUrlExistsForWithCode)
	ctx.Step(`^I view the analytics for the short url with code "([^"]*)"$`, sfh.iViewTheAnalyticsForTheShortUrlWithCode)
	ctx.Step(`^I should receive analytics data for the short url with code "([^"]*)"$`, sfh.iShouldReceiveAnalyticsDataForTheShortUrlWithCode)
	ctx.Step(`^I should receive analytics data for the short url with code "([^"]*)" with (\d+) visit$`, sfh.iShouldReceiveAnalyticsDataForTheShortUrlWithCodeWithVisit)
	ctx.Step(`^I visit the short url with code "([^"]*)"$`, sfh.iVisitTheShortUrlWithCode)
}
