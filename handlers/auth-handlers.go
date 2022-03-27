package handlers

import "net/http"

func (h *Handlers) UserLogin(w http.ResponseWriter, r *http.Request) {
	if err := h.render(w, r, "login", nil, nil); err != nil {
		h.App.ErrorLog.Println(err)
	}
}

func (h *Handlers) PostUserLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	usr, err := h.Models.Users.GetByEmail(email)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	match, err := usr.CheckPassword(password)
	if err != nil {
		_, _ = w.Write([]byte("Error while checking the password"))
		return
	}
	if !match {
		_, _ = w.Write([]byte("Invalid password"))
		return
	}

	h.sessionPut(r.Context(), "userID", usr.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) UserLogout(w http.ResponseWriter, r *http.Request) {
	if err := h.sessionRenew(r.Context()); err != nil {
		h.App.ErrorLog.Println(err)
	}
	h.sessionRemove(r.Context(), "userID")
	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}
