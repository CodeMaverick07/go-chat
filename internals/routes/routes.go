package routes

import (
	"go-chat/internals/app"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	router := chi.NewRouter()

	router.Get("/health", app.HealthCheck)
	router.Route("/auth", func(r chi.Router) {
		r.Post("/register/verify-otp", app.UserHandler.VerifyOTPAndCreateUserHandler)

		r.Post("/otp/send", app.UserHandler.SendOTPHandler)

		r.Post("/login/password", app.AuthHandler.LoginWithEmailOrUsernameAndPassword)
		r.Post("/login/otp", app.AuthHandler.LoginWithEmailandOTP)
		r.Post("/login/otp/verify", app.AuthHandler.VerifyLoginOTP)
	})
	router.Get("/ws", app.WebsocketManager.ServeWS)

	return router
}
