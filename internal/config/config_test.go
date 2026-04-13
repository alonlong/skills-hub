package config

import "testing"

func TestLoad_DefaultHTTPPort(t *testing.T) {
	t.Setenv("BACKEND_HTTP_PORT", "")

	cfg := Load()

	if cfg.HTTPPort != "3001" {
		t.Fatalf("HTTPPort: got %q want %q", cfg.HTTPPort, "3001")
	}
}

func TestLoad_HTTPPortFromEnv(t *testing.T) {
	t.Setenv("BACKEND_HTTP_PORT", "9090")

	cfg := Load()

	if cfg.HTTPPort != "9090" {
		t.Fatalf("HTTPPort: got %q want %q", cfg.HTTPPort, "9090")
	}
}
