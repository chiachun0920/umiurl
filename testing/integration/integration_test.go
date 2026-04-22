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

func SetEnv() func() {
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgres://umiurl:umiurl@localhost:5433/umiurl?sslmode=disable")
	}
	// os.Setenv("SES_SEND_FROM", "noreply@aca5o.com")
	// os.Setenv("FORGET_PASSWORD_ATTEMPT_COUNT", "1")
	return func() {
		os.Unsetenv("DATABASE_URL")
		// os.Unsetenv("SES_SEND_FROM")
		// os.Unsetenv("FORGET_PASSWORD_ATTEMPT_COUNT")
	}

}

func InitializeScenario(ctx *godog.ScenarioContext) {
	shorturl.InitializeShortUrlScenario(ctx, swiss)
}

func InitializeSuite(ctx *godog.TestSuiteContext) {
	var teardown func()
	ctx.BeforeSuite(func() {
		teardown = SetEnv()
		// logger.SetupLogger(utils.FromEnv("LOG_LEVEL"))
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

	ctx.AfterSuite(func() {
		teardown()
	})
}

func TestFeatures(t *testing.T) {
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
