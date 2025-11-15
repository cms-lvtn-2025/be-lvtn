package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadTest loads test-specific configuration
func LoadTest() (*Config, error) {
	// Load test environment file
	testEnvPath := filepath.Join("tests", ".env.test")
	if err := godotenv.Load(testEnvPath); err != nil {
		// If test env file doesn't exist, use defaults
		setTestDefaults()
	}

	return Load()
}

func setTestDefaults() {
	defaults := map[string]string{
		"DB_NAME":                "lvtn_test_db",
		"REDIS_DB":               "1",
		"TLS_ENABLED":            "false",
		"JWT_SECRET":             "test_jwt_secret",
		"LOG_LEVEL":              "debug",
		"TEST_TIMEOUT":           "30s",
		"ENABLE_PLAYGROUND":      "true",
		"ENABLE_INTROSPECTION":   "true",
		"METRICS_ENABLED":        "true",
		"DISABLE_GRAPHQL_AUTH":   "true",
		"SKIP_GRAPHQL_WORKFLOWS": "true",
	}

	for key, value := range defaults {
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}
