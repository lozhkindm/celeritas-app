package main

import (
	"github.com/lozhkindm/celeritas"
	"log"
	"myapp/handlers"
	"os"
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

	return app
}
