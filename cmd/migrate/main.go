package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Sovpalo/sovpalo-backend/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	cfg := config.Load()
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	runner := os.Args
	if len(runner) < 2 {
		runner = append(runner, "up")
	}

	goose.SetDialect("postgres")
	if err := goose.Run(runner[1], db, "migrations", runner[2:]...); err != nil {
		log.Fatalf("goose %s: %v", runner[1], err)
	}
}
