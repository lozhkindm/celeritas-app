package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"

	"myapp/data"
)

func (h *Handlers) UserLogin(w http.ResponseWriter, r *http.Request) {
	if err := h.render(w, r, "login", nil, nil); err != nil {
		h.App.ErrorLog.Println(err)
		h.App.InternalError(w)
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

	if r.Form.Get("remember") == "remember" {
		hasher := sha256.New()
		if _, err := hasher.Write([]byte(h.randomString(12))); err != nil {
			h.App.BadRequest(w)
			return
		}

		var tkn data.RememberToken
		token := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
		if err := tkn.Insert(usr.ID, token); err != nil {
			h.App.BadRequest(w)
			return
		}

		h.App.SetRememberMeCookie(w, usr.ID, token)
		h.sessionPut(r.Context(), "remember_token", token)
	}

	h.sessionPut(r.Context(), "userID", usr.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) UserLogout(w http.ResponseWriter, r *http.Request) {
	if h.sessionHas(r.Context(), "remember_token") {
		var tkn data.RememberToken
		if err := tkn.Delete(h.sessionGetString(r.Context(), "remember_token")); err != nil {
			h.App.ErrorLog.Println(err)
		}
	}
	h.App.DeleteRememberMeCookie(w)

	if err := h.sessionRenew(r.Context()); err != nil {
		h.App.ErrorLog.Println(err)
	}
	h.sessionRemove(r.Context(), "userID")
	h.sessionRemove(r.Context(), "remember_token")
	if err := h.sessionDestroy(r.Context()); err != nil {
		h.App.ErrorLog.Println(err)
	}
	if err := h.sessionRenew(r.Context()); err != nil {
		h.App.ErrorLog.Println(err)
	}
	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}

func (h *Handlers) Forgot(w http.ResponseWriter, r *http.Request) {
	if err := h.render(w, r, "forgot", nil, nil); err != nil {
		h.App.ErrorLog.Println(err)
		h.App.InternalError(w)
	}
}
