package handlers

import (
	"net/http"
	"net/http/httptest"
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
