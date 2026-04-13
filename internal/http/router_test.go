package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	aphttp "skillhub/backend/internal/http"
	"skillhub/backend/internal/config"
)

func TestHealthz(t *testing.T) {
	t.Parallel()
	cfg := config.Load()
	ts := httptest.NewServer(aphttp.NewRouter(cfg, aphttp.Dependencies{}))
	t.Cleanup(ts.Close)

	res, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", res.StatusCode, http.StatusOK)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type: got %q want application/json", ct)
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ok" {
		t.Fatalf("status json: got %q want ok", body.Status)
	}
}
