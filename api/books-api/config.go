package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	AppEnv      string
	DatabaseURL string
}

func LoadConfig() (Config, error) {
	config := Config{
		AppEnv:      os.Getenv("APP_ENV"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
	if config.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if config.AppEnv != "development" && config.AppEnv != "test" && config.AppEnv != "production" {
		return Config{}, fmt.Errorf("APP_ENV must be development, test, or production")
	}
	return config, nil
}

func OpenPool(ctx context.Context, config Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

func RequireTestDatabase(config Config) error {
	if config.AppEnv != "test" {
		return fmt.Errorf("APP_ENV must be test")
	}
	databaseURL, err := url.Parse(config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	if (databaseURL.Scheme != "postgres" && databaseURL.Scheme != "postgresql") || databaseURL.Host == "" {
		return fmt.Errorf("DATABASE_URL must be a PostgreSQL URL")
	}
	if databaseURL.Path != "/writerslife_test" {
		return fmt.Errorf("DATABASE_URL must target /writerslife_test")
	}
	return nil
}
