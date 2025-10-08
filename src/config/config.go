package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Services ServiceConfig
	Google   GoogleOAuthConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type ServiceConfig struct {
	Academic string
	Council  string
	File     string
	Role     string
	Thesis   string
	User     string
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func Load() (*Config, error) {
	if err := godotenv.Load("../../.server.env"); err != nil {
		// Log warning but continue
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8081"),
			Mode: getEnv("GIN_MODE", "release"),
		},
		Services: ServiceConfig{
			Academic: os.Getenv("SERVICE_ACADEMIC"),
			Council:  os.Getenv("SERVICE_COUNCIL"),
			File:     os.Getenv("SERVICE_FILE"),
			Role:     os.Getenv("SERVICE_ROLE"),
			Thesis:   os.Getenv("SERVICE_THESIS"),
			User:     os.Getenv("SERVICE_USER"),
		},
		Google: GoogleOAuthConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
