package routes

import (
	"go-chat/internals/app"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(app *app.Application) *chi.Mux{
	router := chi.NewRouter()
	router.Get("/health",app.HealthCheck)
	router.Post("/create-user",app.UserHandler.CreateUserHandler)
	return router
}
