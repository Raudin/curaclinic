package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"curaclinic/backend/internal/db"
	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/platform/password"
	"curaclinic/backend/internal/repository/postgres"
)

func main() {
	demo := flag.Bool("demo", false, "seed demo doctors, working hours and a patient")
	flag.Parse()
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := db.NewPostgres(ctx, databaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()
	repo := postgres.New(pool)

	if err := seedAdmin(ctx, repo); err != nil {
		log.Fatalf("bootstrap admin: %v", err)
	}
	if *demo {
		if err := seedDemo(ctx, repo); err != nil {
			log.Fatalf("demo data: %v", err)
		}
	}
	fmt.Println("seed completed")
}

func seedAdmin(ctx context.Context, repo *postgres.Repo) error {
	email := strings.ToLower(strings.TrimSpace(envOr("BOOTSTRAP_ADMIN_EMAIL", "")))
	plain := envOr("BOOTSTRAP_ADMIN_PASSWORD", "")
	if email == "" || plain == "" {
		return fmt.Errorf("BOOTSTRAP_ADMIN_EMAIL and BOOTSTRAP_ADMIN_PASSWORD are required")
	}
	if _, err := repo.GetUserByEmail(ctx, email); err == nil {
		fmt.Printf("admin %s already exists; skipping\n", email)
		return nil
	} else if !domain.IsCode(err, domain.CodeNotFound) {
		return err
	}

	role, err := repo.GetRoleByName(ctx, "admin")
	if err != nil {
		return fmt.Errorf("admin role missing; run migrations first: %w", err)
	}
	cost := envInt("BCRYPT_COST", 12)
	if validateErr := password.Validate(plain); validateErr != nil {
		return validateErr
	}
	hash, err := password.Hash(plain, cost)
	if err != nil {
		return err
	}
	userID := uuid.New()
	user := domain.NewUser{
		ID:           userID,
		Email:        email,
		PasswordHash: hash,
		FirstName:    envOr("BOOTSTRAP_ADMIN_FIRST_NAME", "Clinic"),
		LastName:     envOr("BOOTSTRAP_ADMIN_LAST_NAME", "Admin"),
		Status:       domain.UserStatusActive,
	}
	if _, err := repo.CreateUserWithRoles(ctx, user, []uuid.UUID{role.ID}, userID, domain.RequestMeta{IP: "seed"}); err != nil {
		return err
	}
	fmt.Printf("created admin %s\n", email)
	return nil
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
