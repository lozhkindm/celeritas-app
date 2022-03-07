package main

import (
	"log"
	"os"

	"myapp/data"
	"myapp/handlers"

	"github.com/lozhkindm/celeritas"
)

func initApplication() *application {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	cel := &celeritas.Celeritas{}
	if err := cel.New(path); err != nil {
		log.Fatal(err)
	}

	cel.AppName = "myapp"
	cel.Debug = true

	app := &application{
		App:      cel,
		Handlers: &handlers.Handlers{App: cel},
	}

	app.App.Routes = app.routes()
	app.Models = data.New(app.App.DB.Pool)
	app.Handlers.Models = app.Models

	return app
}
