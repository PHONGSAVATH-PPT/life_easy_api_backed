package config

import (
	"errors"
	"net"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, errors.New("cannot load environment file")
	}

	required := []string{
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_DB",
		"JWT_SECRET",
	}

	for _, key := range required {
		if os.Getenv(key) == "" {
			return Config{}, errors.New("missing environment variable: " + key)
		}
	}

	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return Config{}, errors.New("JWT_SECRET must be at least 32 bytes")
	}

	host := value("DB_HOST", "localhost")
	port := value("DB_PORT", "5432")

	dsn := url.URL{
		Scheme: "postgres",
		User: url.UserPassword(
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
		),
		Host: net.JoinHostPort(host, port),
		Path: os.Getenv("POSTGRES_DB"),
	}

	query := dsn.Query()
	query.Set("sslmode", value("DB_SSLMODE", "disable"))
	dsn.RawQuery = query.Encode()

	return Config{
		DatabaseURL: dsn.String(),
		JWTSecret:   secret,
	}, nil
}

func value(key, fallback string) string {
	if result := os.Getenv(key); result != "" {
		return result
	}
	return fallback
}
