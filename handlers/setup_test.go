package handlers

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/lozhkindm/celeritas"
	"github.com/lozhkindm/celeritas/mailer"
	"github.com/lozhkindm/celeritas/render"
)

var (
	cel          celeritas.Celeritas
	testSession  *scs.SessionManager
	testHandlers Handlers
)

func TestMain(m *testing.M) {
	infoLog := log.New(os.Stdout, "INFO", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR", log.Ldate|log.Ltime|log.Lshortfile)

	testSession = scs.New()
	testSession.Lifetime = 24 * time.Hour
	testSession.Cookie.Persist = true
	testSession.Cookie.SameSite = http.SameSiteLaxMode
	testSession.Cookie.Secure = false

	var views = jet.NewSet(jet.NewOSFileSystemLoader("../views"), jet.InDevelopmentMode())

	renderer := render.Render{
		Renderer: "jet",
		RootPath: "../",
		Port:     "4000",
		JetViews: views,
		Session:  testSession,
	}

	cel = celeritas.Celeritas{
		AppName:       "myapp",
		Debug:         true,
		Version:       "1.0.0",
		ErrorLog:      errorLog,
		InfoLog:       infoLog,
		RootPath:      "../",
		Routes:        nil,
		Render:        &renderer,
		Session:       testSession,
		DB:            celeritas.Database{},
		JetViews:      views,
		EncryptionKey: cel.RandStr(32),
		Cache:         nil,
		Scheduler:     nil,
		Mail:          mailer.Mail{},
		Server:        celeritas.Server{},
	}

	testHandlers.App = &cel

	os.Exit(m.Run())
}
