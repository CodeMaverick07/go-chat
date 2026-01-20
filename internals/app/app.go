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
	Logger                     *log.Logger
	DB                         *sql.DB
	UserHandler                *api.UserHandler
	EmailSender                *email.Sender
	TokenHandler               *api.TokenHandler
	AuthHandler                *api.AuthHandler
	MessageHandler             *api.MessageHandler
	ConversationHandler        *api.ConversationHandler
	UserMiddlewareHandler      middleware.UserMiddleware
	WebsocketManager           *websockets.Manager
	WebSocketMiddlewareHandler middleware.WebsocketMiddleware
	MediaHandler               *api.MediaHandler
}

func NewApplication() (*Application, error) {
	// Logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Database
	db, err := store.Open()
	if err != nil {
		return nil, err
	}

	if err := store.MigrateFS(db, migrations.FS, "."); err != nil {
		return nil, err
	}

	// External / Infra Services
	emailCfg := email.LoadConfig()
	emailSender := email.NewSender(
		emailCfg.Host,
		emailCfg.Port,
		emailCfg.Username,
		emailCfg.Password,
	)

	// Stores (Repositories)
	userStore := store.NewUserStore(db)
	otpStore := store.NewOTPStore(db, emailSender)
	tokenStore := store.NewPostgresTokenStore(db)
	conversationStore := store.NewPostgresConversationStore(db)
	messageStore := store.NewPostgresMessageStore(db)

	// Handlers (API Layer)
	userHandler := api.NewUserHandler(
		userStore,
		logger,
		otpStore,
		tokenStore,
	)

	authHandler := api.NewAuthHandler(
		logger,
		userStore,
		tokenStore,
		otpStore,
	)

	tokenHandler := api.NewTokenHandler(
		tokenStore,
		userStore,
		logger,
	)

	conversationHandler := api.NewConversationHandler(
		messageStore,
		conversationStore,
		logger,
	)

	messageHandler := api.NewMessageHandler(
		messageStore,
		conversationStore,
		logger,
	)
	mediaHandler := api.NewMediaHandler(messageHandler, conversationHandler, logger)

	// Middleware
	userMiddlewareHandler := middleware.UserMiddleware{
		UserStore: userStore,
	}

	websocketMiddlewareHandler := middleware.WebsocketMiddleware{
		UserStore: userStore,
	}
	websocketManager := websockets.NewManager(logger, messageHandler, conversationHandler)

	// Application
	return &Application{
		Logger:           logger,
		DB:               db,
		EmailSender:      emailSender,
		WebsocketManager: websocketManager,

		UserHandler:         userHandler,
		AuthHandler:         authHandler,
		TokenHandler:        tokenHandler,
		ConversationHandler: conversationHandler,
		MessageHandler:      messageHandler,

		UserMiddlewareHandler:      userMiddlewareHandler,
		WebSocketMiddlewareHandler: websocketMiddlewareHandler,
		MediaHandler:               mediaHandler,
	}, nil
}

func (a *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "server is running successfully")
}
