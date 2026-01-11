package app

import (
	"database/sql"
	"fmt"
	"go-chat/internals/store"
	"log"
	"net/http"
	"os"
)


type Application struct {
	Logger *log.Logger
	DB     *sql.DB
}

func NewApplication ()(*Application,error){
	logger := log.New(os.Stdout,"",log.Ldate|log.Ltime)
	db,err := store.Open()
	if err != nil {
		panic(err)
	}
	
	return &Application{
		Logger: logger,
		DB: db,
	},nil
}

func (a *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "server is running successfully")
}