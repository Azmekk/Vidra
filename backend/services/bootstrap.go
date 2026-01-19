package services

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Bootstrap(ctx context.Context) (*pgxpool.Pool, string) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = "postgres://postgres:postgres@localhost:5432/vidra?sslmode=disable"
	}

	// Parse the URL to get the database name and a connection string to the 'postgres' database
	config, err := pgx.ParseConfig(dbUrl)
	if err != nil {
		log.Fatalf("Unable to parse DATABASE_URL: %v", err)
	}
	targetDB := config.Database
	config.Database = "postgres"
	postgresURL := config.ConnString()

	// Connect to 'postgres' database to ensure the target database exists
	pgConn, err := pgx.Connect(ctx, postgresURL)
	if err != nil {
		log.Fatalf("Unable to connect to postgres database: %v", err)
	}

	var exists bool
	err = pgConn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", targetDB).Scan(&exists)
	if err != nil {
		log.Fatalf("Error checking if database exists: %v", err)
	}

	if !exists {
		log.Printf("Database %s does not exist, creating it...\n", targetDB)
		_, err = pgConn.Exec(ctx, "CREATE DATABASE "+targetDB)
		if err != nil {
			log.Fatalf("Error creating database: %v", err)
		}
		log.Printf("Database %s created successfully.\n", targetDB)
	}
	pgConn.Close(ctx)

	// Ensure downloads directory exists
	if err := os.MkdirAll("downloads", 0755); err != nil {
		log.Fatalf("Error creating downloads directory: %v", err)
	}

	// Run migrations
	runMigrations(dbUrl)

	// Now connect to the target database with pgxpool
	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database pool: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return pool, port
}

func runMigrations(dbUrl string) {
	m, err := migrate.New(
		"file://sql/migrations",
		dbUrl,
	)
	if err != nil {
		log.Fatalf("could not create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("could not run up migrations: %v", err)
	}

	log.Println("Migrations applied successfully!")
}
