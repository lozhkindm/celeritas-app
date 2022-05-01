package celeritas

import (
	"fmt"
	"net/http"
	"regexp"
	"runtime"
	"time"
)

func (c *Celeritas) LoadTime(start time.Time) {
	elapsed := time.Since(start)
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	regexpFn := regexp.MustCompile(`^.*\.(.*)$`)
	caller := regexpFn.ReplaceAllString(fn.Name(), "$1")
	c.InfoLog.Printf("Load Time: %q took %s\n", caller, elapsed)
}

func (c *Celeritas) SetRememberMeCookie(w http.ResponseWriter, ID int, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     c.GetRememberMeCookieName(),
		Value:    fmt.Sprintf("%d|%s", ID, token),
		Path:     "/",
		Domain:   c.Session.Cookie.Domain,
		Expires:  time.Now().Add(365 * 24 * 60 * 60 * time.Second),
		MaxAge:   315360000,
		Secure:   c.Session.Cookie.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (c *Celeritas) DeleteRememberMeCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     c.GetRememberMeCookieName(),
		Value:    "",
		Path:     "/",
		Domain:   c.Session.Cookie.Domain,
		Expires:  time.Now().Add(-100 * time.Hour),
		MaxAge:   -1,
		Secure:   c.Session.Cookie.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (c *Celeritas) GetRememberMeCookieName() string {
	return fmt.Sprintf("_%s_remember", c.AppName)
}
