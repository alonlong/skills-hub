package config

import "os"

// Config holds process configuration loaded from the environment.
type Config struct {
	HTTPPort               string
	DatabaseURL            string
	JWTSecret              string
	StorageBasePath        string
	BootstrapAdminEnabled  bool
	BootstrapAdminUserID   string
	BootstrapAdminUsername string
	BootstrapAdminPassword string
	BootstrapAdminName     string
	BootstrapAdminEmail    string
}

// Load reads configuration from the environment.
// BACKEND_HTTP_PORT defaults to 3001 when unset or empty.
func Load() Config {
	port := os.Getenv("BACKEND_HTTP_PORT")
	if port == "" {
		port = "3001"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://skillhub:skillhub_dev@localhost:5432/skillhub?sslmode=disable"
	}

	jwtSecret := os.Getenv("BACKEND_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret"
	}

	storageBasePath := os.Getenv("STORAGE_BASE_PATH")
	if storageBasePath == "" {
		storageBasePath = "/tmp/skillhub-storage"
	}

	return Config{
		HTTPPort:               port,
		DatabaseURL:            databaseURL,
		JWTSecret:              jwtSecret,
		StorageBasePath:        storageBasePath,
		BootstrapAdminEnabled:  os.Getenv("BOOTSTRAP_ADMIN_ENABLED") == "true",
		BootstrapAdminUserID:   getenvDefault("BOOTSTRAP_ADMIN_USER_ID", "00000000-0000-0000-0000-000000000001"),
		BootstrapAdminUsername: getenvDefault("BOOTSTRAP_ADMIN_USERNAME", "admin"),
		BootstrapAdminPassword: getenvDefault("BOOTSTRAP_ADMIN_PASSWORD", "123456"),
		BootstrapAdminName:     getenvDefault("BOOTSTRAP_ADMIN_DISPLAY_NAME", "Admin"),
		BootstrapAdminEmail:    getenvDefault("BOOTSTRAP_ADMIN_EMAIL", "admin@skillhub.local"),
	}
}

func getenvDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
