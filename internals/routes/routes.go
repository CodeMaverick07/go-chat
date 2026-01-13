package routes

import (
	"go-chat/internals/app"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	router := chi.NewRouter()
	router.Get("/health", app.HealthCheck)
	router.Post("/verify-otp-and-register-user", app.UserHandler.VerifyOTPAndCreateUserHandler)
	router.Post("/send-otp", app.UserHandler.SendOTPHandler)

	return router
}
