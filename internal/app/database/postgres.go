package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"project/internal/app/config"
)

func NewPostgres(ctx context.Context, cfg config.PostgresConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn(cfg))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.PingContext(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("ping postgres: %w; close postgres: %v", err, closeErr)
		}

		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

func dsn(cfg config.PostgresConfig) string {
	sslMode := "disable"
	if cfg.SSL == "true" {
		sslMode = "require"
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		sslMode,
	)
}
