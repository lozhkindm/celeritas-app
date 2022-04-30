package middlewares

import (
	"fmt"
	"net/http"
)

func (m *Middleware) CheckRemember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.App.Session.Exists(r.Context(), "userID") {
			next.ServeHTTP(w, r)
		} else {
			_, err := r.Cookie(fmt.Sprintf("_%s_remember", m.App.AppName))
			if err != nil {
				next.ServeHTTP(w, r)
			}

		}
	})
}
