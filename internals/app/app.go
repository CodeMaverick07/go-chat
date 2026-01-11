package app

import (
	"database/sql"
	"fmt"
	"go-chat/internals/api"
	"go-chat/internals/store"
	"go-chat/migrations"
	"log"
	"net/http"
	"os"
)


type Application struct {
	Logger *log.Logger
	DB     *sql.DB
	UserHandler *api.UserHandler
}

func NewApplication ()(*Application,error){
	logger := log.New(os.Stdout,"",log.Ldate|log.Ltime)
	db,err := store.Open()
	if err != nil {
		panic(err)
	}
	err = store.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	userStore := store.NewUserStore(db)
	newUserHandler:= api.NewUserHandler(userStore,logger)
	return &Application{
		Logger: logger,
		DB: db,
		UserHandler :newUserHandler,
	},nil
}

func (a *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "server is running successfully")
}