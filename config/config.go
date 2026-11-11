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
		return Config{}, errors.New("cannot load .env")
	}

	if os.Getenv("JWT_SECRET") == "" {
		return Config{}, errors.New("missing environment: JWT_SECRET")
	}

	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		if err := validateDatabaseURL(dsn); err != nil {
			return Config{}, err
		}

		return Config{
			DatabaseURL: dsn,
			JWTSecret:   os.Getenv("JWT_SECRET"),
		}, nil
	}

	for _, key := range []string{
		"POSTGRES_DB",
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
	} {
		if os.Getenv(key) == "" {
			return Config{}, errors.New("missing environment: " + key)
		}
	}

	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return Config{}, errors.New("JWT_SECRET must contain at least 32 bytes")
	}

	dsn := url.URL{
		Scheme: "postgres",
		User: url.UserPassword(
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
		),
		Host: net.JoinHostPort(
			value("DB_HOST", "localhost"),
			value("DB_PORT", "5432"),
		),
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

func validateDatabaseURL(dsn string) error {
	parsed, err := url.Parse(dsn)
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") || parsed.Host == "" || parsed.Path == "" {
		return errors.New("DATABASE_URL must be a valid PostgreSQL URL")
	}

	return nil
}

func value(key, fallback string) string {
	if result := os.Getenv(key); result != "" {
		return result
	}
	return fallback
}
