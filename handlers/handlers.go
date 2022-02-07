package handlers

import (
	"github.com/CloudyKit/jet/v6"
	"github.com/lozhkindm/celeritas"
	"myapp/data"
	"net/http"
)

type Handlers struct {
	App    *celeritas.Celeritas
	Models data.Models
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	if err := h.App.Render.Page(w, r, "home", nil, nil); err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) GoPage(w http.ResponseWriter, r *http.Request) {
	h.App.Render.Renderer = "go"
	if err := h.App.Render.Page(w, r, "home", nil, nil); err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) JetPage(w http.ResponseWriter, r *http.Request) {
	h.App.Render.Renderer = "jet"
	if err := h.App.Render.Page(w, r, "jet-template", nil, nil); err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) Sessions(w http.ResponseWriter, r *http.Request) {
	myData := "bar"
	h.App.Session.Put(r.Context(), "foo", myData)
	myValue := h.App.Session.GetString(r.Context(), "foo")
	vars := make(jet.VarMap)
	vars.Set("foo", myValue)

	if err := h.App.Render.Page(w, r, "sessions", vars, nil); err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}
