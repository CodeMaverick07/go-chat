package main

import (
	"go-chat/internals/app"
	"go-chat/internals/routes"
	"net/http"
	"time"
)


func main() {
	app,err := app.NewApplication()
	if err != nil {
		panic(err)
	}
	r:=routes.SetupRoutes(app)
	defer app.DB.Close()
	app.Logger.Println("first log from main.go")
	server := &http.Server{
		Addr: ":9000",
		Handler: r,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	app.Logger.Println("starting server on port 9000")
	err = server.ListenAndServe()
	if err != nil {
		app.Logger.Fatal("error starting server: %w", err)
	}
	
}