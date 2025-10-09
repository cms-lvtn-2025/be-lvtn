package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Services ServiceMap
	Google   GoogleOAuthConfig
	Redis    RedisConfig
	MongoDB  MongoConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type ServiceConfig struct {
	Port    string
	Endpont string
}

type MinioConfig struct {
	Endpont    string
	AccessKey  string
	SecretKey  string
	UseSSL     bool
	BucketName string
}

type ServiceMap struct {
	Academic ServiceConfig
	Council  ServiceConfig
	File     ServiceConfig
	Role     ServiceConfig
	Thesis   ServiceConfig
	User     ServiceConfig
	MinIo    MinioConfig
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type RedisConfig struct {
	Address      string
	Password     string
	DB           int
	DialTimeout  int // seconds
	ReadTimeout  int // seconds
	WriteTimeout int // seconds
	PoolSize     int
	MinIdleConns int
}

type MongoConfig struct {
	URI                     string
	Database                string
	Username                string
	Password                string
	AuthSource              string
	MaxPoolSize             int
	MinPoolSize             int
	MaxConnIdleTime         int // seconds
	ConnectTimeout          int // seconds
	ServerSelectionTimeout  int // seconds
}

type JWTConfig struct {
	AccessSecret         string
	RefreshSecret        string
	AccessTokenExpiry    int // minutes
	RefreshTokenExpiry   int // days
}

func Load() (*Config, error) {
	// Try multiple paths for .server.env
	envPaths := []string{
		"env/.server.env",           // From project root
		"../env/.server.env",        // From src/
		"../../env/.server.env",     // From src/config/
		".server.env",               // Current directory
	}

	loaded := false
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			loaded = true
			break
		}
	}

	if !loaded {
		// Log warning but continue with environment variables
		fmt.Println("Warning: Could not load .server.env file, using environment variables")
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8081"),
			Mode: getEnv("GIN_MODE", "release"),
		},
		Services: ServiceMap{
			Academic: ServiceConfig{
				Port:    getEnv("PORT", "500051"),
				Endpont: getEnv("SERVICE_ACADEMIC_URL", "academic"),
			},
			Council: ServiceConfig{
				Port:    getEnv("PORT", "500052"),
				Endpont: getEnv("SERVICE_COUNCIL_URL", "council"),
			},
			File: ServiceConfig{
				Port:    getEnv("PORT", "500053"),
				Endpont: getEnv("SERVICE_FILE_URL", "file"),
			},
			Role: ServiceConfig{
				Port:    getEnv("PORT", "500054"),
				Endpont: getEnv("SERVICE_ROLE_URL", "role"),
			},
			Thesis: ServiceConfig{
				Port:    getEnv("PORT", "500055"),
				Endpont: getEnv("SERVICE_THESIS_URL", "thesis"),
			},
			User: ServiceConfig{
				Port:    getEnv("PORT", "500056"),
				Endpont: getEnv("SERVICE_USER_URL", "user"),
			},
			MinIo: MinioConfig{
				Endpont:    os.Getenv("MIN_ENDPOINT"),
				AccessKey:  os.Getenv("MIN_ACCESS_KEY"),
				SecretKey:  os.Getenv("MIN_SECRET_KEY"),
				UseSSL:     getEnv("MIN_USE_SSL", "false") == "true",
				BucketName: os.Getenv("MIN_BUCKET_NAME"),
			},
		},
		Google: GoogleOAuthConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: getEnv("GOOGLE_SECRET", os.Getenv("GOOGLE_CLIENT_SECRET")),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECTURL"),
		},
		Redis: RedisConfig{
			Address:      getEnv("REDIS_ADDRESS", "localhost:6379"),
			Password:     os.Getenv("REDIS_PASSWORD"),
			DB:           getEnvAsInt("REDIS_DB", 0),
			DialTimeout:  getEnvAsInt("REDIS_DIAL_TIMEOUT", 5),
			ReadTimeout:  getEnvAsInt("REDIS_READ_TIMEOUT", 3),
			WriteTimeout: getEnvAsInt("REDIS_WRITE_TIMEOUT", 3),
			PoolSize:     getEnvAsInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 2),
		},
		MongoDB: MongoConfig{
			URI:                    getEnv("MONGO_URI", "mongodb://localhost:27017"),
			Database:               getEnv("MONGO_DATABASE", "thesis_db"),
			Username:               os.Getenv("MONGO_USERNAME"),
			Password:               os.Getenv("MONGO_PASSWORD"),
			AuthSource:             getEnv("MONGO_AUTH_SOURCE", "admin"),
			MaxPoolSize:            getEnvAsInt("MONGO_MAX_POOL_SIZE", 100),
			MinPoolSize:            getEnvAsInt("MONGO_MIN_POOL_SIZE", 10),
			MaxConnIdleTime:        getEnvAsInt("MONGO_MAX_CONN_IDLE_TIME", 60),
			ConnectTimeout:         getEnvAsInt("MONGO_CONNECT_TIMEOUT", 10),
			ServerSelectionTimeout: getEnvAsInt("MONGO_SERVER_SELECTION_TIMEOUT", 10),
		},
		JWT: JWTConfig{
			AccessSecret:       getEnv("JWT_ACCESS_SECRET", "your-secret-key-change-this-in-production"),
			RefreshSecret:      getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key-change-this-in-production"),
			AccessTokenExpiry:  getEnvAsInt("JWT_ACCESS_EXPIRY", 15),  // 15 minutes
			RefreshTokenExpiry: getEnvAsInt("JWT_REFRESH_EXPIRY", 7), // 7 days
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

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	var value int
	if _, err := fmt.Sscanf(valueStr, "%d", &value); err != nil {
		return defaultValue
	}
	return value
}
