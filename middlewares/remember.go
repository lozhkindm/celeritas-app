package middlewares

import (
	"net/http"
	"strconv"
	"strings"

	"myapp/data"
)

func (m *Middleware) CheckRemember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.App.Session.Exists(r.Context(), "userID") {
			if cookie, err := r.Cookie(m.App.GetRememberMeCookieName()); err == nil {
				cv := cookie.Value
				if len(cv) > 0 {
					var u data.User
					parts := strings.Split(cv, "|")
					uid, hash := parts[0], parts[1]
					userID, err := strconv.Atoi(uid)
					if err != nil {
						m.App.ErrorLog.Println(err)
					}
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

func (m *Middleware) deleteRememberCookie(w http.ResponseWriter, r *http.Request) {
	if err := m.App.Session.RenewToken(r.Context()); err != nil {
		m.App.ErrorLog.Println(err)
	}
	m.App.DeleteRememberMeCookie(w)
	m.App.Session.Remove(r.Context(), "userID")
	if err := m.App.Session.Destroy(r.Context()); err != nil {
		m.App.ErrorLog.Println(err)
	}
	if err := m.App.Session.RenewToken(r.Context()); err != nil {
		m.App.ErrorLog.Println(err)
	}
}
