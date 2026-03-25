package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName      string
	Environment  string
	HTTP         HTTPConfig
	Database     DatabaseConfig
	Redis        RedisConfig
	Kafka        KafkaConfig
	Auth         AuthConfig
	Integrations IntegrationsConfig
}

type HTTPConfig struct {
	Address         string
	FrontendDistDir string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
}

type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Enabled      bool
	Addr         string
	Password     string
	DB           int
	DashboardTTL time.Duration
}

type KafkaConfig struct {
	Enabled      bool
	Brokers      []string
	ClientID     string
	AuditTopic   string
	WriteTimeout time.Duration
}

type AuthConfig struct {
	JWTSecret               string
	JWTIssuer               string
	AccessTokenTTL          time.Duration
	RefreshTokenTTL         time.Duration
	AllowInsecureDevAuth    bool
	DefaultBaseCurrency     string
	DefaultTimezone         string
	DefaultWeeklyReviewHour int
}

type IntegrationsConfig struct {
	Telegram TelegramConfig
	Yandex   YandexConfig
}

type TelegramConfig struct {
	BotToken string
}

type YandexConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}

	return cfg
}

func Load() (*Config, error) {
	cfg := &Config{
		AppName:     getEnv("APP_NAME", "moneyapp-backend"),
		Environment: getEnv("APP_ENV", "development"),
		HTTP: HTTPConfig{
			Address:         getEnv("HTTP_ADDR", ":8080"),
			FrontendDistDir: getEnv("FRONTEND_DIST_DIR", ""),
			ReadTimeout:     getDurationEnv("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getDurationEnv("HTTP_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     getDurationEnv("HTTP_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			DSN:             getEnv("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/moneyapp?sslmode=disable"),
			MaxOpenConns:    getIntEnv("DATABASE_MAX_OPEN_CONNS", 20),
			MaxIdleConns:    getIntEnv("DATABASE_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getDurationEnv("DATABASE_CONN_MAX_LIFETIME", 30*time.Minute),
		},
		Redis: RedisConfig{
			Enabled:      getBoolEnv("REDIS_ENABLED", true),
			Addr:         getEnv("REDIS_ADDR", "localhost:6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getIntEnv("REDIS_DB", 0),
			DashboardTTL: getDurationEnv("REDIS_DASHBOARD_TTL", 30*time.Second),
		},
		Kafka: KafkaConfig{
			Enabled:      getBoolEnv("KAFKA_ENABLED", true),
			Brokers:      getSliceEnv("KAFKA_BROKERS", []string{"localhost:9092"}),
			ClientID:     getEnv("KAFKA_CLIENT_ID", "moneyapp-backend"),
			AuditTopic:   getEnv("KAFKA_AUDIT_TOPIC", "moneyapp.audit"),
			WriteTimeout: getDurationEnv("KAFKA_WRITE_TIMEOUT", 5*time.Second),
		},
		Auth: AuthConfig{
			JWTSecret:               getEnv("AUTH_JWT_SECRET", "change-me"),
			JWTIssuer:               getEnv("AUTH_JWT_ISSUER", "moneyapp"),
			AccessTokenTTL:          getDurationEnv("AUTH_ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL:         getDurationEnv("AUTH_REFRESH_TOKEN_TTL", 30*24*time.Hour),
			AllowInsecureDevAuth:    getBoolEnv("AUTH_ALLOW_INSECURE_DEV_AUTH", true),
			DefaultBaseCurrency:     getEnv("DEFAULT_BASE_CURRENCY", "RUB"),
			DefaultTimezone:         getEnv("DEFAULT_TIMEZONE", "Europe/Amsterdam"),
			DefaultWeeklyReviewHour: getIntEnv("DEFAULT_WEEKLY_REVIEW_HOUR", 18),
		},
		Integrations: IntegrationsConfig{
			Telegram: TelegramConfig{
				BotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
			},
			Yandex: YandexConfig{
				ClientID:     getEnv("YANDEX_CLIENT_ID", ""),
				ClientSecret: getEnv("YANDEX_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("YANDEX_REDIRECT_URL", ""),
			},
		},
	}

	if cfg.Auth.JWTSecret == "change-me" && cfg.Environment == "production" {
		return nil, fmt.Errorf("AUTH_JWT_SECRET must be configured in production")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return fallback
}

func getIntEnv(key string, fallback int) int {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getBoolEnv(key string, fallback bool) bool {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getSliceEnv(key string, fallback []string) []string {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			result = append(result, s)
		}
	}

	return result
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
