package main

import (
	"fmt"
	"net/http"
	"strconv"

	"myapp/data"

	"github.com/go-chi/chi/v5"
)

func (a *application) routes() *chi.Mux {
	// middlewares

	// routes
	a.App.Routes.Get("/", a.Handlers.Home)
	a.App.Routes.Get("/go-page", a.Handlers.GoPage)
	a.App.Routes.Get("/jet-page", a.Handlers.JetPage)
	a.App.Routes.Get("/sessions", a.Handlers.Sessions)
	a.App.Routes.Get("/users/login", a.Handlers.UserLogin)
	a.App.Routes.Post("/users/login", a.Handlers.PostUserLogin)
	a.App.Routes.Get("/users/logout", a.Handlers.UserLogout)

	a.App.Routes.Get("/create-user", func(w http.ResponseWriter, r *http.Request) {
		usr := data.User{
			FirstName: a.App.RandStr(10),
			LastName:  a.App.RandStr(10),
			Email:     a.App.RandStr(10),
			Active:    1,
			Password:  "password",
		}
		id, err := a.Models.Users.Insert(&usr)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}
		fmt.Fprintf(w, "%d: %s", id, usr.FirstName)
	})
	a.App.Routes.Get("/get-all-users", func(w http.ResponseWriter, r *http.Request) {
		users, err := a.Models.Users.GetAll()
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}
		for _, u := range users {
			fmt.Fprintf(w, "%+v", u)
		}
	})
	a.App.Routes.Get("/get-user/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))
		user, err := a.Models.Users.GetById(id)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}
		fmt.Fprintf(w, "%+v", user)
	})
	a.App.Routes.Get("/update-user/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))
		user, err := a.Models.Users.GetById(id)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		user.LastName = a.App.RandStr(10)
		validator := a.App.Validator(nil)
		validator.Check(len(user.LastName) > 20, "last_name", "minimum 20 chars")
		if !validator.IsValid() {
			fmt.Fprint(w, "failed validation")
			return
		}

		if err := a.Models.Users.Update(user); err != nil {
			a.App.ErrorLog.Println(err)
			return
		}
		fmt.Fprintf(w, "%+v", user)
	})

	// static routes
	fileServer := http.FileServer(http.Dir("./public"))
	a.App.Routes.Handle("/public/*", http.StripPrefix("/public", fileServer))

	return a.App.Routes
}
