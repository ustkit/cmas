package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ustkit/cmas/internal/server/config"
	"github.com/ustkit/cmas/internal/server/handlers"
	"github.com/ustkit/cmas/internal/types"
)

func NewRouter(serverConfig *config.Config, repo types.MetricRepo) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Compress(5))

	h := handlers.NewHandler(serverConfig, repo)

	r.Get("/", h.Index)

	r.Get("/ping", h.Ping)

	r.Route("/update", func(r chi.Router) {
		r.Post("/", h.UpdateJSON)
		r.Post("/{type}/{name}/{value}", h.UpdatePlain)
	})

	r.Route("/updates", func(r chi.Router) {
		r.Post("/", h.UpdateJSONBatch)
	})

	r.Route("/value", func(r chi.Router) {
		r.Post("/", h.ValueJSON)
		r.Get("/{type}/{name}", h.ValuePlain)
	})

	return r
}
