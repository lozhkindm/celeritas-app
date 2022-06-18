package handlers

import (
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
