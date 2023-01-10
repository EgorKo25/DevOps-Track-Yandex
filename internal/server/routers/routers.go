package routers

import (
	"github.com/EgorKo25/DevOps-Track-Yandex/internal/server/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(handler *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.GetAllStats)
		r.Get("/ping", handler.PingDB)
		r.Get("/value/{type}/{name}", handler.GetValueStat)
		r.Post("/update/{type}/{name}/{value}", handler.SetMetricValue)

		r.Post("/updates/", handler.GetJSONUpdates)
		r.Post("/update/", handler.SetJSONValue)
		r.Post("/value/", handler.GetJSONValue)

	})

	return r
}
