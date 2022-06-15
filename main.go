package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"myapp/data"
	"myapp/handlers"
	"myapp/middlewares"

	"github.com/lozhkindm/celeritas"
)

type application struct {
	App         *celeritas.Celeritas
	Handlers    *handlers.Handlers
	Models      data.Models
	Middlewares *middlewares.Middleware
	wg          sync.WaitGroup
}

func main() {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		log.Fatalf("failed to load location: %s", err)
	}
	time.Local = loc

	c := initApplication()
	go c.listenForShutdown()
	if err := c.App.ListenAndServe(); err != nil {
		c.App.ErrorLog.Println(err)
	}
}

func (a *application) shutdown() {
	// clean up tasks
	a.wg.Wait()
}

func (a *application) listenForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit
	a.App.InfoLog.Println("Received signal:", s.String())
	a.shutdown()
	os.Exit(0)
}
