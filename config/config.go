package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const minJWTSecretBytes = 32

type Config struct {
	AppName            string
	AppPort            string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBSSLMode          string
	JWTSecret          string
	JWTExpiresIn       time.Duration
	CORSAllowedOrigins []string
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("gagal membaca file .env: %w", err)
	}

	cfg := Config{
		AppName:            os.Getenv("APP_NAME"),
		AppPort:            envOr("APP_PORT", "8080"),
		CORSAllowedOrigins: splitList(os.Getenv("CORS_ALLOWED_ORIGINS")),
	}
	if err := loadDB(&cfg); err != nil {
		return Config{}, err
	}
	if err := loadJWT(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func loadDB(cfg *Config) error {
	cfg.DBHost = os.Getenv("DB_HOST")
	cfg.DBPort = os.Getenv("DB_PORT")
	cfg.DBUser = os.Getenv("DB_USER")
	cfg.DBPassword = os.Getenv("DB_PASSWORD")
	cfg.DBName = os.Getenv("DB_NAME")
	cfg.DBSSLMode = envOr("DB_SSLMODE", "disable")

	if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBUser == "" || cfg.DBName == "" {
		return fmt.Errorf("konfigurasi database belum lengkap")
	}
	return nil
}

func loadJWT(cfg *Config) error {
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if len(cfg.JWTSecret) < minJWTSecretBytes {
		return fmt.Errorf("JWT_SECRET minimal %d byte", minJWTSecretBytes)
	}

	ttl, err := time.ParseDuration(os.Getenv("JWT_EXPIRES_IN"))
	if err != nil || ttl <= 0 {
		return fmt.Errorf("JWT_EXPIRES_IN tidak valid, contoh: 8h")
	}

	cfg.JWTExpiresIn = ttl
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitList(raw string) []string {
	var out []string
	for item := range strings.SplitSeq(raw, ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}
