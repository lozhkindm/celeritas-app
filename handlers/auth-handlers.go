package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"myapp/data"

	"github.com/CloudyKit/jet/v6"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/lozhkindm/celeritas/mailer"
	"github.com/lozhkindm/celeritas/urlsigner"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
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

func (h *Handlers) PostForgot(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.App.BadRequest(w)
		return
	}

	var u *data.User
	email := r.Form.Get("email")
	u, err := u.GetByEmail(email)
	if err != nil {
		h.App.BadRequest(w)
		return
	}

	link := fmt.Sprintf("%s/users/reset-password?email=%s", h.App.Server.URL, email)
	signer := urlsigner.Signer{Secret: []byte(h.App.EncryptionKey)}
	signedLink := signer.GenerateTokenFromString(link)

	var msgData struct {
		Link string
	}
	msgData.Link = signedLink
	msg := mailer.Message{
		From:     "senka@ignatov.com",
		FromName: "Senka",
		To:       u.Email,
		Subject:  "Reset password",
		Template: "reset-password",
		Data:     msgData,
	}

	h.App.Mail.Jobs <- msg
	res := <-h.App.Mail.Results
	if res.Error != nil {
		h.App.InternalError(w)
		return
	}

	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}

func (h *Handlers) ResetPasswordForm(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	url := fmt.Sprintf("%s%s", h.App.Server.URL, r.RequestURI)

	signer := urlsigner.Signer{Secret: []byte(h.App.EncryptionKey)}
	if !signer.VerifyToken(url) {
		h.App.Unauthorized(w)
		return
	}
	if signer.Expired(url, 60*time.Minute) {
		h.App.Unauthorized(w)
		return
	}

	cryptoEmail, err := h.encrypt(email)
	if err != nil {
		h.App.InternalError(w)
		return
	}

	vars := make(jet.VarMap)
	vars.Set("email", cryptoEmail)
	if err := h.render(w, r, "reset-password", vars, nil); err != nil {
		h.App.InternalError(w)
		return
	}
}

func (h *Handlers) PostResetPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.App.InternalError(w)
		return
	}

	email, err := h.decrypt(r.Form.Get("email"))
	if err != nil {
		h.App.InternalError(w)
		return
	}

	var u *data.User
	u, err = u.GetByEmail(email)
	if err != nil {
		h.App.InternalError(w)
		return
	}

	if err := u.ResetPassword(u.ID, r.Form.Get("password")); err != nil {
		h.App.InternalError(w)
		return
	}

	h.App.Session.Put(r.Context(), "flash", "Password reset. You can now log in.")
	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}

func (h *Handlers) InitSocialAuth() {
	scope := []string{"user"}
	goth.UseProviders(
		github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), os.Getenv("GITHUB_CALLBACK"), scope...),
	)
	st := sessions.NewCookieStore([]byte(os.Getenv("KEY")))
	st.MaxAge(86400 * 30)
	st.Options.Path = "/"
	st.Options.HttpOnly = true
	st.Options.Secure = false
	gothic.Store = st
}

func (h *Handlers) SocialLogin(w http.ResponseWriter, r *http.Request) {
	h.App.Session.Put(r.Context(), "social_provider", chi.URLParam(r, "provider"))
	h.InitSocialAuth()
	if _, err := gothic.CompleteUserAuth(w, r); err != nil {
		gothic.BeginAuthHandler(w, r)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (h *Handlers) SocialMediaCallback(w http.ResponseWriter, r *http.Request) {
	h.InitSocialAuth()
	gUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		h.App.Session.Put(r.Context(), "error", err.Error())
		http.Redirect(w, r, "/users/login", http.StatusSeeOther)
		return
	}

	user, err := h.Models.Users.GetByEmail(gUser.Email)
	if err != nil {
		user = &data.User{
			Email:     gUser.Email,
			Active:    1,
			Password:  h.App.RandStr(20),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		provider := h.App.Session.Get(r.Context(), "social_provider").(string)
		switch provider {
		case "github":
			partsName := strings.Split(gUser.Name, " ")
			user.FirstName = partsName[0]
			if len(partsName) > 1 {
				user.LastName = partsName[1]
			}
		case "google":
		}

		if _, err := h.Models.Users.Insert(user); err != nil {
			h.App.InternalError(w)
			return
		}
	}

	h.App.Session.Put(r.Context(), "userID", user.ID)
	h.App.Session.Put(r.Context(), "social_token", gUser.AccessToken)
	h.App.Session.Put(r.Context(), "social_email", gUser.Email)
	h.App.Session.Put(r.Context(), "flash", "Successfully logged in")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
