package handlers

import (
	"github.com/CloudyKit/jet/v6"
	"myapp/data"
	"net/http"
)

func (h *Handlers) ShowForm(w http.ResponseWriter, r *http.Request) {
	vars := make(jet.VarMap)
	validator := h.App.Validator(nil)
	vars.Set("validator", validator)
	vars.Set("user", data.User{})

	if err := h.App.Render.Page(w, r, "form", vars, nil); err != nil {
		h.App.ErrorLog.Println(err)
	}
}
