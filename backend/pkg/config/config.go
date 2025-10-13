package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	ServerPort string
	WSPort     string
	WebsiteURL string

	// Database
	DBURL string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// RabbitMQ
	RabbitMQURL             string
	RabbitMQExchange        string
	RabbitMQShardingEnabled bool

	// JWT
	JWTSecret        string
	JWTResetSecret   string
	JWTExpiration    time.Duration

	// Storage (MinIO / S3-compatible)
	StorageEndpoint  string
	StorageAccessKey string
	StorageSecretKey string
	StorageBucket    string
	StorageUseSSL    bool
	StorageBaseURL   string

	// OpenSearch
	OpenSearchHost     string
	OpenSearchUsername string
	OpenSearchPassword string
	SearchEnabled      bool
	SearchIndexPrefix  string
	SearchReindexOnStartup bool

	// Email (Resend)
	ResendAPIKey   string
	ResendFromEmail string

	// Web Push (VAPID)
	WebPushPublicKey  string
	WebPushPrivateKey string
	WebPushSubscriber string

	// OAuth
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string

	// App settings
	UploadMaxSizeMB   int64
	SnippetLength     int
	SiteName          string
	SiteDescription   string
}

func Load() (*Config, error) {
	// Load .env file if present (non-fatal if missing)
	if err := godotenv.Load(); err != nil {
		slog.Info("no .env file found, using environment variables")
	}

	cfg := &Config{
		ServerPort:              getEnv("SERVER_PORT", "8080"),
		WSPort:                  getEnv("WS_PORT", "8082"),
		WebsiteURL:              getEnv("WEBSITE_URL", "http://localhost:3000"),
		RabbitMQExchange:        getEnv("RABBITMQ_EXCHANGE", "orbit-exchange"),
		RabbitMQShardingEnabled: getEnvBool("RABBITMQ_SHARDING_ENABLED", true),
		StorageEndpoint:         getEnv("STORAGE_ENDPOINT", "localhost:9000"),
		StorageAccessKey:        getEnv("STORAGE_ACCESS_KEY", "minioadmin"),
		StorageSecretKey:        getEnv("STORAGE_SECRET_KEY", "minioadmin"),
		StorageBucket:           getEnv("STORAGE_BUCKET", "orbit"),
		StorageUseSSL:           getEnvBool("STORAGE_USE_SSL", false),
		StorageBaseURL:          getEnv("STORAGE_BASE_URL", "http://localhost:9000/orbit"),
		OpenSearchHost:          getEnv("OPENSEARCH_HOST", "http://localhost:9200"),
		OpenSearchUsername:      getEnv("OPENSEARCH_USERNAME", ""),
		OpenSearchPassword:      getEnv("OPENSEARCH_PASSWORD", ""),
		SearchEnabled:           getEnvBool("SEARCH_ENABLED", true),
		SearchIndexPrefix:       getEnv("SEARCH_INDEX_PREFIX", "orbit"),
		SearchReindexOnStartup:  getEnvBool("SEARCH_REINDEX_ON_STARTUP", false),
		ResendAPIKey:            getEnv("RESEND_API_KEY", ""),
		ResendFromEmail:         getEnv("RESEND_FROM_EMAIL", "noreply@example.com"),
		WebPushPublicKey:        getEnv("WEBPUSH_PUBLIC_KEY", ""),
		WebPushPrivateKey:       getEnv("WEBPUSH_PRIVATE_KEY", ""),
		WebPushSubscriber:       getEnv("WEBPUSH_SUBSCRIBER", "mailto:admin@example.com"),
		GoogleClientID:          getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:      getEnv("GOOGLE_CLIENT_SECRET", ""),
		GitHubClientID:          getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret:      getEnv("GITHUB_CLIENT_SECRET", ""),
		UploadMaxSizeMB:         getEnvInt64("UPLOAD_MAX_SIZE_MB", 10),
		SnippetLength:           getEnvInt("SNIPPET_LENGTH", 200),
		SiteName:                getEnv("SITE_NAME", "Orbit"),
		SiteDescription:         getEnv("SITE_DESCRIPTION", "A community forum"),
	}

	// Build DB URL
	cfg.DBURL = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		getEnv("DB_USER", "orbit"),
		getEnv("DB_PASSWORD", "orbit"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "orbit"),
		getEnv("DB_SSL_MODE", "disable"),
	)

	// Build Redis addr
	cfg.RedisAddr = fmt.Sprintf("%s:%s",
		getEnv("REDIS_HOST", "localhost"),
		getEnv("REDIS_PORT", "6379"),
	)
	cfg.RedisPassword = getEnv("REDIS_PASSWORD", "")
	cfg.RedisDB = getEnvInt("REDIS_DB", 0)

	// Build RabbitMQ URL
	cfg.RabbitMQURL = fmt.Sprintf("amqp://%s:%s@%s:%s/",
		getEnv("RABBITMQ_USERNAME", "guest"),
		getEnv("RABBITMQ_PASSWORD", "guest"),
		getEnv("RABBITMQ_HOST", "localhost"),
		getEnv("RABBITMQ_PORT", "5672"),
	)

	// JWT
	cfg.JWTSecret = getEnv("JWT_SECRET", "changeme_jwt_secret")
	cfg.JWTResetSecret = getEnv("JWT_RESET_SECRET", "changeme_reset_secret")
	expirationHours := getEnvInt("JWT_EXPIRATION_HOURS", 720)
	cfg.JWTExpiration = time.Duration(expirationHours) * time.Hour

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func getEnvInt64(key string, fallback int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fallback
	}
	return i
}
