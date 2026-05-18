package main

import (
	"context"
	"time"

	"project/internal/app/config"
	"project/internal/app/database"
	"project/internal/app/logger"
)

var migrateLog = logger.New("migrate")

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.NewPostgres(ctx, cfg.Postgres)
	if err != nil {
		migrateLog.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(ctx, db); err != nil {
		migrateLog.Fatalf("run migrations: %v", err)
	}

	migrateLog.Successf("migrations applied")
}
