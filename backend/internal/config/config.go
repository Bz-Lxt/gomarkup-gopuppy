package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv             string
	LogLevel           string
	HTTPAddr           string
	JWTSecret          string
	AccessTTL          time.Duration
	RefreshTTL         time.Duration
	DatabaseURL        string
	RedisURL           string
	CORSOrigins        []string
	StorageDriver      string
	StorageLocalRoot   string
	OSSEndpoint        string
	OSSBucket          string
	OSSAccessKeyID     string
	OSSAccessKeySecret string
	COSBucketURL       string
	COSSecretID        string
	COSSecretKey       string
	SMTPHost           string
	SMTPPort           int
	SMTPFrom           string
	SMTPUser           string
	SMTPPass           string
	WecomWebhookURL    string
	GenericWebhookURL  string
	NotifierMode       string
	ReminderScanHour   int
	ReminderTickSec    int
}

func Load() (*Config, error) {
	c := &Config{
		AppEnv:             env("APP_ENV", "development"),
		LogLevel:           env("LOG_LEVEL", "info"),
		HTTPAddr:           env("HTTP_ADDR", ":8080"),
		JWTSecret:          env("JWT_SECRET", ""),
		DatabaseURL:        env("DATABASE_URL", ""),
		RedisURL:           env("REDIS_URL", "redis://redis:6379/0"),
		StorageDriver:      strings.ToLower(env("STORAGE_DRIVER", "local")),
		StorageLocalRoot:   env("STORAGE_LOCAL_ROOT", "/data/uploads"),
		OSSEndpoint:        env("OSS_ENDPOINT", ""),
		OSSBucket:          env("OSS_BUCKET", ""),
		OSSAccessKeyID:     env("OSS_ACCESS_KEY_ID", ""),
		OSSAccessKeySecret: env("OSS_ACCESS_KEY_SECRET", ""),
		COSBucketURL:       env("COS_BUCKET_URL", ""),
		COSSecretID:        env("COS_SECRET_ID", ""),
		COSSecretKey:       env("COS_SECRET_KEY", ""),
		SMTPHost:           env("SMTP_HOST", "mailhog"),
		SMTPFrom:           env("SMTP_FROM", "gopuppy@local.test"),
		SMTPUser:           env("SMTP_USER", ""),
		SMTPPass:           env("SMTP_PASS", ""),
		WecomWebhookURL:    env("WECOM_WEBHOOK_URL", ""),
		GenericWebhookURL:  env("GENERIC_WEBHOOK_URL", ""),
		NotifierMode:       strings.ToLower(env("NOTIFIER_MODE", "mock")),
		ReminderScanHour:   envInt("REMINDER_SCAN_HOUR", 8),
		ReminderTickSec:    envInt("REMINDER_TICK_SECONDS", 30),
		SMTPPort:           envInt("SMTP_PORT", 1025),
	}
	var err error
	c.AccessTTL, err = time.ParseDuration(env("JWT_ACCESS_TTL", "2h"))
	if err != nil {
		return nil, fmt.Errorf("JWT_ACCESS_TTL: %w", err)
	}
	c.RefreshTTL, err = time.ParseDuration(env("JWT_REFRESH_TTL", "168h"))
	if err != nil {
		return nil, fmt.Errorf("JWT_REFRESH_TTL: %w", err)
	}
	if c.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	for _, o := range strings.Split(env("CORS_ORIGINS", ""), ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			c.CORSOrigins = append(c.CORSOrigins, o)
		}
	}
	if len(c.CORSOrigins) == 0 {
		return nil, fmt.Errorf("CORS_ORIGINS must be an explicit whitelist")
	}
	for _, o := range c.CORSOrigins {
		if o == "*" {
			return nil, fmt.Errorf("CORS_ORIGINS must not contain *")
		}
	}
	return c, nil
}

func env(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
