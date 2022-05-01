package middlewares

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"myapp/data"
)

func (m *Middleware) CheckRemember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.App.Session.Exists(r.Context(), "userID") {
			if cookie, err := r.Cookie(m.getRememberMeCookieName()); err == nil {
				cv := cookie.Value
				if len(cv) > 0 {
					var u data.User
					parts := strings.Split(cv, "|")
					uid, hash := parts[0], parts[1]
					userID, _ := strconv.Atoi(uid)
					if u.CheckRememberToken(userID, hash) {
						if _, err := u.GetById(userID); err == nil {
							m.App.Session.Put(r.Context(), "userID", userID)
							m.App.Session.Put(r.Context(), "remember_token", hash)
						}
					} else {
						m.deleteRememberCookie(w, r)
						m.App.Session.Put(r.Context(), "error", "You have been logged out from another device")
					}
				} else {
					m.deleteRememberCookie(w, r)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) getRememberMeCookieName() string {
	return fmt.Sprintf("_%s_remember", m.App.AppName)
}

func (m *Middleware) deleteRememberCookie(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())
	http.SetCookie(w, &http.Cookie{
		Name:     m.getRememberMeCookieName(),
		Value:    "",
		Path:     "/",
		Domain:   m.App.Session.Cookie.Domain,
		Expires:  time.Now().Add(100 * time.Hour),
		MaxAge:   -1,
		Secure:   m.App.Session.Cookie.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	m.App.Session.Remove(r.Context(), "userID")
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())
}
