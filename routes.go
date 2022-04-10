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
	a.routeGet("/", a.Handlers.Home)
	a.routeGet("/go-page", a.Handlers.GoPage)
	a.routeGet("/jet-page", a.Handlers.JetPage)
	a.routeGet("/sessions", a.Handlers.Sessions)
	a.routeGet("/users/login", a.Handlers.UserLogin)
	a.routePost("/users/login", a.Handlers.PostUserLogin)
	a.routeGet("/users/logout", a.Handlers.UserLogout)
	a.routeGet("/form", a.Handlers.ShowForm)
	a.routePost("/form", a.Handlers.SubmitForm)
	a.routeGet("/json", a.Handlers.JSON)
	a.routeGet("/xml", a.Handlers.XML)
	a.routeGet("/download-file", a.Handlers.DownloadFile)
	a.routeGet("/crypto", a.Handlers.TestCrypto)
	a.routeGet("/cache-test", a.Handlers.ShowCachePage)
	a.routePost("/api/save-in-cache", a.Handlers.SaveInCache)
	a.routePost("/api/get-from-cache", a.Handlers.GetFromCache)
	a.routePost("/api/delete-from-cache", a.Handlers.DeleteFromCache)
	a.routePost("/api/empty-cache", a.Handlers.EmptyCache)

	a.routeGet("/create-user", func(w http.ResponseWriter, r *http.Request) {
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
	a.routeGet("/get-all-users", func(w http.ResponseWriter, r *http.Request) {
		users, err := a.Models.Users.GetAll()
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}
		for _, u := range users {
			fmt.Fprintf(w, "%+v", u)
		}
	})
	a.routeGet("/get-user/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))
		user, err := a.Models.Users.GetById(id)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}
		fmt.Fprintf(w, "%+v", user)
	})
	a.routeGet("/update-user/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))
		user, err := a.Models.Users.GetById(id)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		user.LastName = a.App.RandStr(10)
		validator := a.App.Validator(nil)
		user.FirstName = ""
		user.Validate(validator)
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
