package handlers

import (
	"fmt"
	"net/http"

	"myapp/data"

	"github.com/CloudyKit/jet/v6"
	"github.com/lozhkindm/celeritas"
)

type Handlers struct {
	App    *celeritas.Celeritas
	Models data.Models
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	if err := h.render(w, r, "home", nil, nil); err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) GoPage(w http.ResponseWriter, r *http.Request) {
	h.App.Render.Renderer = "go"
	if err := h.render(w, r, "home", nil, nil); err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) JetPage(w http.ResponseWriter, r *http.Request) {
	h.App.Render.Renderer = "jet"
	if err := h.render(w, r, "jet-template", nil, nil); err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) Sessions(w http.ResponseWriter, r *http.Request) {
	myData := "bar"
	h.sessionPut(r.Context(), "foo", myData)
	myValue := h.App.Session.GetString(r.Context(), "foo")
	vars := make(jet.VarMap)
	vars.Set("foo", myValue)

	if err := h.render(w, r, "sessions", vars, nil); err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) JSON(w http.ResponseWriter, _ *http.Request) {
	var payload struct {
		ID      int64    `json:"id"`
		Name    string   `json:"name"`
		Hobbies []string `json:"hobbies"`
	}

	payload.ID = 123
	payload.Name = "Ignat Senkin"
	payload.Hobbies = []string{"CS", "Formula", "Cards"}

	if err := h.App.WriteJSON(w, http.StatusOK, payload); err != nil {
		h.App.ErrorLog.Println(err)
	}
}

func (h *Handlers) XML(w http.ResponseWriter, _ *http.Request) {
	type payload struct {
		ID      int64    `xml:"id"`
		Name    string   `xml:"name"`
		Hobbies []string `xml:"hobbies>hobby"`
	}
	var pl payload

	pl.ID = 123
	pl.Name = "Ignat Senkin"
	pl.Hobbies = []string{"CS", "Formula", "Cards"}

	if err := h.App.WriteXML(w, http.StatusOK, pl); err != nil {
		h.App.ErrorLog.Println(err)
	}
}

func (h *Handlers) DownloadFile(w http.ResponseWriter, r *http.Request) {
	h.App.DownloadFile(w, r, "./public/images", "celeritas.jpg")
}

func (h *Handlers) TestCrypto(w http.ResponseWriter, _ *http.Request) {
	plaintext := "Hello, world"
	fmt.Fprintf(w, "plaintext: %s\n", plaintext)

	encrypted, err := h.encrypt(plaintext)
	if err != nil {
		h.App.ErrorLog.Println(err)
		h.App.InternalError(w)
		return
	}

	fmt.Fprintf(w, "encrypted: %s\n", encrypted)

	decrypted, err := h.decrypt(encrypted)
	if err != nil {
		h.App.ErrorLog.Println(err)
		h.App.InternalError(w)
		return
	}

	fmt.Fprintf(w, "decrypted: %s", decrypted)
}
