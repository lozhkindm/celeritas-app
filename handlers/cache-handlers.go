package handlers

import "net/http"

func (h *Handlers) ShowCachePage(w http.ResponseWriter, r *http.Request) {
	if err := h.render(w, r, "cache", nil, nil); err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) SaveInCache(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name  string `json:"name"`
		Value string `json:"value"`
		CSRF  string `json:"csrf_token"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.InternalError(w)
		return
	}

	if err := h.App.Cache.Set(input.Name, input.Value); err != nil {
		h.App.InternalError(w)
		return
	}

	var res struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	res.Error = false
	res.Message = "Saved in cache"

	if err := h.App.WriteJSON(w, http.StatusCreated, res); err != nil {
		h.App.InternalError(w)
		return
	}
}

func (h *Handlers) GetFromCache(w http.ResponseWriter, r *http.Request) {

}

func (h *Handlers) DeleteFromCache(w http.ResponseWriter, r *http.Request) {

}

func (h *Handlers) EmptyCache(w http.ResponseWriter, r *http.Request) {

}
