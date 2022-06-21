package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlers_Home(t *testing.T) {
	server := httptest.NewTLSServer(getRoutes())
	defer server.Close()

	res, err := server.Client().Get(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Errorf("expected status 200, bot %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "awesome") {
		if err := cel.TakeScreenshot(server.URL, "Home", 1500, 1000); err != nil {
			t.Errorf("failed to take a screenshot: %s", err)
		}
		t.Error("failed to find awesome on the page")
	}
}

func TestHandlers_Home2(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Errorf("failed to create a request: %s", err)
	}

	r = r.WithContext(getCtx(r))
	w := httptest.NewRecorder()

	cel.Session.Put(r.Context(), "test_key", "Hello")
	h := http.HandlerFunc(testHandlers.Home)
	h.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected status 200, bot %d", w.Code)
	}
	if cel.Session.GetString(r.Context(), "test_key") != "Hello" {
		t.Error("got wrong value from session")
	}
}

func TestHandlers_Clicker(t *testing.T) {
	server := httptest.NewTLSServer(getRoutes())
	defer server.Close()

	page := cel.GetPage(server.URL + "/tester")
	output := cel.GetElementByID(page, "output")
	button := cel.GetElementByID(page, "clicker")

	html, err := output.HTML()
	if err != nil {
		t.Error(err)
	}
	if html != `<div id="output"></div>` {
		t.Error("output is not empty")
	}

	button.MustClick()
	html, err = output.HTML()
	if err != nil {
		t.Error(err)
	}
	if html != `<div id="output">Clicked the button!</div>` {
		t.Error("output is empty")
	}
}
