package routes

import (
	"go-chat/internals/app"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://127.0.0.1:5500",
			"http://localhost:5500",
		},
		AllowedMethods: []string{
			"GET", "POST", "OPTIONS",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
		},
		ExposedHeaders: []string{
			"Link",
		},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Protected routes (require authentication)
	router.Group(func(r chi.Router) {
		r.Use(app.UserMiddlewareHandler.Authenticate)
		r.Post("/socket-token", app.UserHandler.WebsocketTokenHandler)

		// Media upload endpoint
		r.Post("/upload", app.MediaHandler.UploadMedia)
	})

	// WebSocket endpoint
	router.Group(func(r chi.Router) {
		r.Use(app.WebSocketMiddlewareHandler.AuthenticateWebsockets)
		r.Get("/ws", app.WebsocketManager.ServeWS)
	})

	// Serve uploaded media files
	router.Get("/uploads/*", app.MediaHandler.ServeMedia)

	// Health check
	router.Get("/health", app.HealthCheck)

	// Auth routes
	router.Route("/auth", func(r chi.Router) {
		r.Post("/register/verify-otp", app.UserHandler.VerifyOTPAndCreateUserHandler)
		r.Post("/otp/send", app.UserHandler.SendOTPHandler)
		r.Post("/login/password", app.AuthHandler.LoginWithEmailOrUsernameAndPassword)
		r.Post("/login/otp", app.AuthHandler.LoginWithEmailandOTP)
		r.Post("/login/otp/verify", app.AuthHandler.VerifyLoginOTP)
	})

	return router
}
