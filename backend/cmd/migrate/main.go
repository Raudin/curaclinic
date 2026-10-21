// Command migrate applies or rolls back the embedded database migrations.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"

	"curaclinic/backend/migrations"
)

func main() {
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	if err := sqlDB.PingContext(context.Background()); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	goose.SetBaseFS(migrations.FS)
	goose.SetSequential(true)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("configure goose: %v", err)
	}

	args := os.Args[2:]
	if err := goose.RunContext(context.Background(), command, sqlDB, ".", args...); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	fmt.Printf("migrations %s completed\n", command)
}
