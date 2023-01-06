package registry

import (
	"context"
	"net/http"

	"github.com/distribution/distribution/v3/configuration"
	"github.com/distribution/distribution/v3/health"
	"github.com/distribution/distribution/v3/registry/handlers"
)

func NewHTTPHandler(config *Config) http.Handler {
	ctx := context.Background()
	app := handlers.NewApp(ctx, config)
	app.RegisterHealthChecks()
	handler := health.Handler(app)
	return handler
}

type Config = configuration.Configuration
