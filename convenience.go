package main

import "net/http"

func (a *application) routeGet(s string, h http.HandlerFunc) {
	a.App.Routes.Get(s, h)
}

func (a *application) routePost(s string, h http.HandlerFunc) {
	a.App.Routes.Post(s, h)
}

func (a *application) routeUse(m ...func(http.Handler) http.Handler) {
	a.App.Routes.Use(m...)
}
