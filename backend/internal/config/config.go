package config

import (
	"fmt"
	"net/url"
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
	Auth         AuthConfig
	Features     FeatureConfig
	Integrations IntegrationsConfig
}

type HTTPConfig struct {
	Address         string
	FrontendDistDir string
	UploadsDir      string
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

type FeatureConfig struct {
	CourseIntakeManagerApprovalEnabled bool
}

type IntegrationsConfig struct {
	Telegram TelegramConfig
	Yandex   YandexConfig
	Outlook  OutlookConfig
	YandexAI YandexAIConfig
}

type YandexAIConfig struct {
	APIKey   string
	FolderID string
	Model    string
}

type TelegramConfig struct {
	BotToken string
}

type YandexConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type OutlookConfig struct {
	TenantID       string
	ClientID       string
	ClientSecret   string
	RedirectURI    string
	PostConnectURL string
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
			UploadsDir:      getEnv("HTTP_UPLOADS_DIR", "uploads"),
			ReadTimeout:     getDurationEnv("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getDurationEnv("HTTP_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     getDurationEnv("HTTP_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			DSN:             getDatabaseDSN(),
			MaxOpenConns:    getIntEnv("DATABASE_MAX_OPEN_CONNS", 20),
			MaxIdleConns:    getIntEnv("DATABASE_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getDurationEnv("DATABASE_CONN_MAX_LIFETIME", 30*time.Minute),
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
		Features: FeatureConfig{
			CourseIntakeManagerApprovalEnabled: getBoolEnv("FEATURE_COURSE_INTAKE_MANAGER_APPROVAL", false),
		},
		Integrations: IntegrationsConfig{
			Telegram: TelegramConfig{
				BotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
			},
			Yandex: YandexConfig{
				ClientID:     getEnv("YANDEX_CLIENT_ID", ""),
				ClientSecret: getEnv("YANDEX_CLIENT_SECRET", ""),
				RedirectURI:  getEnv("YANDEX_REDIRECT_URI", firstNonEmptyEnv("VITE_YANDEX_REDIRECT_URI", "")),
			},
			Outlook: OutlookConfig{
				TenantID:       getEnv("OUTLOOK_TENANT_ID", "common"),
				ClientID:       getEnv("OUTLOOK_CLIENT_ID", ""),
				ClientSecret:   getEnv("OUTLOOK_CLIENT_SECRET", ""),
				RedirectURI:    getEnv("OUTLOOK_REDIRECT_URI", ""),
				PostConnectURL: getEnv("OUTLOOK_POST_CONNECT_URL", "/calendar/overview"),
			},
			YandexAI: YandexAIConfig{
				APIKey:   getEnv("YANDEX_AI_API_KEY", ""),
				FolderID: getEnv("YANDEX_AI_FOLDER_ID", "b1gste4lfr39is20f5r8"),
				Model:    getEnv("YANDEX_AI_MODEL", "gpt-oss-20b/latest"),
			},
		},
	}

	if cfg.Auth.JWTSecret == "change-me" && cfg.Environment == "production" {
		return nil, fmt.Errorf("AUTH_JWT_SECRET must be configured in production")
	}

	return cfg, nil
}

func getDatabaseDSN() string {
	if dsn := getEnv("DATABASE_DSN", ""); dsn != "" {
		return dsn
	}

	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	database := getEnv("POSTGRES_DB", "moneyapp")
	user := getEnv("POSTGRES_USER", "postgres")
	password := getEnv("POSTGRES_PASSWORD", "postgres")

	return (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     host + ":" + port,
		Path:     database,
		RawQuery: "sslmode=disable",
	}).String()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return fallback
}

func firstNonEmptyEnv(key, fallback string) string {
	value := strings.TrimSpace(getEnv(key, ""))
	if value != "" {
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
