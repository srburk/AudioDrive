package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"audiodrive/internal/server"
	"audiodrive/internal/store"
)

func TestRoutes_Registered(t *testing.T) {
	s := store.NewInMemory()
	srv := server.New(":0", s)

	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	routes := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodPost, "/urls", http.StatusUnprocessableEntity}, // no body → 422
		{http.MethodGet, "/urls", http.StatusOK},
		{http.MethodGet, "/urls/999", http.StatusNotFound},
	}

	for _, rt := range routes {
		req, _ := http.NewRequest(rt.method, ts.URL+rt.path, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("%s %s: %v", rt.method, rt.path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != rt.want {
			t.Errorf("%s %s: status = %d, want %d", rt.method, rt.path, resp.StatusCode, rt.want)
		}
	}
}
