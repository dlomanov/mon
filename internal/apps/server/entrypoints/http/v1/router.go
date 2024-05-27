package v1

import (
	"github.com/dlomanov/mon/internal/apps/server/container"
	"github.com/dlomanov/mon/internal/apps/server/entrypoints/http/middlewares"
	_ "github.com/dlomanov/mon/internal/apps/server/entrypoints/http/v1/docs"
	"github.com/dlomanov/mon/internal/apps/server/entrypoints/http/v1/endpoints"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// UseEndpoints - wires the endpoints with the provided router.
// It also sets up the middleware.
//
//	@title		mon API
//	@version	1.0
func UseEndpoints(r chi.Router, c *container.Container) {
	logger := c.Logger
	r.Use(middleware.Recoverer)
	r.Use(middlewares.Logger(logger))
	r.Use(middlewares.TrustedSubnet(logger, c.Config.TrustedSubnet))
	r.Use(middlewares.Compressor)
	r.Use(middlewares.Decrypter(logger, c.Dec))
	r.Use(middlewares.Hash(c))
	endpoints.UseSwagger(r, c)
	endpoints.UseMetrics(r, c)
	r.Get("/ping", endpoints.PingDB(c))
}
