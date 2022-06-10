package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (a *application) ApiRoutes() http.Handler {
	return chi.NewRouter().Route("/api", func(r chi.Router) {
		r.Get("/test-api", func(w http.ResponseWriter, r *http.Request) {
			var payload struct {
				Content string `json:"content"`
			}
			payload.Content = "This is content from api route!"
			_ = a.App.WriteJSON(w, http.StatusOK, payload)
		})
	})
}
