package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const minJWTSecretBytes = 32

type Config struct {
	AppName      string
	AppPort      string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	JWTSecret    string
	JWTExpiresIn time.Duration
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("gagal membaca file .env: %w", err)
	}

	cfg := Config{
		AppName:    os.Getenv("APP_NAME"),
		AppPort:    os.Getenv("APP_PORT"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSSLMode:  os.Getenv("DB_SSLMODE"),
	}

	if cfg.AppPort == "" {
		cfg.AppPort = "8080"
	}

	if cfg.DBSSLMode == "" {
		cfg.DBSSLMode = "disable"
	}

	if cfg.DBHost == "" || cfg.DBPort == "" ||
		cfg.DBUser == "" || cfg.DBName == "" {
		return Config{}, fmt.Errorf("konfigurasi database belum lengkap")
	}

	if err := loadJWT(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func loadJWT(cfg *Config) error {
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if len(cfg.JWTSecret) < minJWTSecretBytes {
		return fmt.Errorf("JWT_SECRET minimal %d byte", minJWTSecretBytes)
	}

	ttl, err := time.ParseDuration(os.Getenv("JWT_EXPIRES_IN"))
	if err != nil || ttl <= 0 {
		return fmt.Errorf("JWT_EXPIRES_IN tidak valid, contoh: 24h")
	}

	cfg.JWTExpiresIn = ttl
	return nil
}
