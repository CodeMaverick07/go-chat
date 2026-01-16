package app

import (
	"database/sql"
	"fmt"
	"go-chat/internals/api"
	"go-chat/internals/email"
	"go-chat/internals/middleware"
	"go-chat/internals/store"
	"go-chat/internals/websockets"
	"go-chat/migrations"
	"log"
	"net/http"
	"os"
)

type Application struct {
	Logger            *log.Logger
	DB                *sql.DB
	UserHandler       *api.UserHandler
	EmailSender       *email.Sender
	TokenHandler      *api.TokenHandler
	AuthHandler       *api.AuthHandler
	MiddlewareHandler middleware.UserMiddleware
	WebsocketManager  *websockets.Manager
}

func NewApplication() (*Application, error) {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	db, err := store.Open()
	if err != nil {
		panic(err)
	}
	err = store.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}
	emailCfg := email.LoadConfig()
	emailSender := email.NewSender(emailCfg.Host, emailCfg.Port, emailCfg.Username, emailCfg.Password)

	userStore := store.NewUserStore(db)
	otpStore := store.NewOTPStore(db, emailSender)
	tokenStore := store.NewPostgresTokenStore(db)
	userHandler := api.NewUserHandler(userStore, logger, otpStore, tokenStore)
	authHandler := api.NewAuthHandler(logger, userStore, tokenStore, otpStore)
	tokenHander := api.NewTokenHandler(tokenStore, userStore, logger)
	middlewareHandler := middleware.UserMiddleware{UserStore: userStore}
	websocketManger := websockets.NewManager(logger)
	return &Application{
		Logger:            logger,
		DB:                db,
		UserHandler:       userHandler,
		EmailSender:       emailSender,
		TokenHandler:      tokenHander,
		AuthHandler:       authHandler,
		MiddlewareHandler: middlewareHandler,
		WebsocketManager:  websocketManger,
	}, nil
}

func (a *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "server is running successfully")
}
