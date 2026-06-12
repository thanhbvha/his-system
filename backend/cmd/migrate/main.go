package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load() // Ignore error if .env doesn't exist

	dbURL := os.Getenv("POSTGRES_DSN")
	if dbURL == "" {
		log.Fatal("POSTGRES_DSN environment variable is required")
	}

	upFlag := flag.Bool("up", false, "Run up migrations")
	downFlag := flag.Bool("down", false, "Run down migrations")
	seedFlag := flag.Bool("seed", false, "Run seed data")

	flag.Parse()

	// If no flags are provided, default to --up
	if !*upFlag && !*downFlag && !*seedFlag {
		*upFlag = true
		*seedFlag = true
	}

	m, err := migrate.New("file://migrations/postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	if *upFlag {
		err = m.Up()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("Migration up applied successfully!")
	} else if *downFlag {
		err = m.Down()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("Migration down applied successfully!")
	}

	if *seedFlag {
		ctx := context.Background()
		conn, err := pgx.Connect(ctx, dbURL)
		if err != nil {
			log.Fatalf("Unable to connect to database for seeding: %v\n", err)
		}
		defer conn.Close(ctx)

		seedSQL, err := os.ReadFile("migrations/postgres/seed/001_seed.sql")
		if err != nil {
			log.Fatalf("Failed to read seed file: %v", err)
		}

		_, err = conn.Exec(ctx, string(seedSQL))
		if err != nil {
			log.Fatalf("Failed to execute seed SQL: %v", err)
		}
		fmt.Println("Seed data applied successfully!")
	}
}
