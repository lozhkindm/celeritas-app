package main

import (
	"log"
	"time"

	"myapp/data"
	"myapp/handlers"

	"github.com/lozhkindm/celeritas"
)

type application struct {
	App      *celeritas.Celeritas
	Handlers *handlers.Handlers
	Models   data.Models
}

func main() {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		log.Fatalf("failed to load location: %s", err)
	}
	time.Local = loc

	c := initApplication()
	c.App.ListenAndServe()
}
