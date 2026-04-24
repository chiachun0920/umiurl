package integration

import (
	"context"
	"os"
	"testing"

	"umiurl/testing/integration/shorturl"
	"umiurl/testing/integration/testswiss"

	"github.com/cucumber/godog"
)

type ContextKey string

var swiss *testswiss.TestSwiss

func InitializeScenario(ctx *godog.ScenarioContext) {
	shorturl.InitializeShortUrlScenario(ctx, swiss)
}

func InitializeSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		swiss = testswiss.NewTestSwiss()
		resetDb()
	})

	ctx.ScenarioContext().Before(
		func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
			truncateDb()
			return ctx, nil
		},
	)

	ctx.ScenarioContext().After(
		func(ctx context.Context, scenario *godog.Scenario, err error) (context.Context, error) {
			return ctx, nil
		},
	)
}

func TestFeatures(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("integration tests are disabled; set RUN_INTEGRATION_TESTS=1 and TEST_DATABASE_URL to enable them")
	}

	suite := godog.TestSuite{
		TestSuiteInitializer: InitializeSuite,
		ScenarioInitializer:  InitializeScenario,
		Options: &godog.Options{
			Format: "pretty",
			Paths: []string{
				"shorturl/features",
			},
			TestingT: t,
			// Tags:     "debug",
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
