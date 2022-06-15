package celeritas

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func (c *Celeritas) ListenAndServe() error {
	defer func() {
		if c.DB.Pool != nil {
			_ = c.DB.Pool.Close()
		}
		if c.redisPool != nil {
			_ = c.redisPool.Close()
		}
		if c.badgerConn != nil {
			_ = c.badgerConn.Close()
		}
	}()
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler:      c.Routes,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
		IdleTimeout:  30 * time.Second,
		ErrorLog:     c.ErrorLog,
	}
	go c.listenRPC()
	c.InfoLog.Printf("Listening on port %s", os.Getenv("PORT"))
	return srv.ListenAndServe()
}
