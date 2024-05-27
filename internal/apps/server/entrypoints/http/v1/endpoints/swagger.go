package endpoints

import (
	"github.com/dlomanov/mon/internal/apps/server/container"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"strings"
)

func UseSwagger(r chi.Router, c *container.Container) {
	url := c.Config.Addr
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(url+"/swagger/doc.json"),
	))
}
