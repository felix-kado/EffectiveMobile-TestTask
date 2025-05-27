package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"person-api/internal/services/person"
)

// NewRouter на chi
func NewRouter(svc person.Service) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// swagger
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("doc.json"), // swagger endpoint
	))

	// CRUD /persons
	r.Route("/persons", func(r chi.Router) {
		r.Get("/", handleList(svc))
		r.Post("/", handleCreate(svc))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handleGetByID(svc))
			r.Put("/", handleUpdate(svc))
			r.Delete("/", handleDelete(svc))
		})
	})

	return r
}
