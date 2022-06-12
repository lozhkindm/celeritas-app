package celeritas

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/justinas/nosurf"
)

func (c *Celeritas) SessionLoad(next http.Handler) http.Handler {
	return c.Session.LoadAndSave(next)
}

func (c *Celeritas) CSRFToken(next http.Handler) http.Handler {
	handler := nosurf.New(next)
	secure, err := strconv.ParseBool(c.config.cookie.secure)
	if err != nil {
		c.ErrorLog.Fatal(err)
	}

	handler.ExemptGlob("/api/*")

	handler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Domain:   c.config.cookie.domain,
	})
	return handler
}

func (c *Celeritas) MaintenanceMode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if maintenanceMode && !strings.Contains(r.URL.Path, "public") && !strings.Contains(r.URL.Path, "maintenance.html") {
			c.ResponseMaintenanceMode(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
