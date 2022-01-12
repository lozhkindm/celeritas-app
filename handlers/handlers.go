package handlers

import (
	"github.com/lozhkindm/celeritas"
	"net/http"
)

type Handlers struct {
	App *celeritas.Celeritas
}

func (h Handlers) Home(w http.ResponseWriter, r *http.Request) {
	if err := h.App.Render.Page(w, r, "home", nil, nil); err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}
