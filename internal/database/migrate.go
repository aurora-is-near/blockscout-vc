package database

import (
	"database/sql"
	"embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// isValidDatabaseName validates database names to prevent SQL injection
func isValidDatabaseName(name string) error {
	if len(name) < 1 || len(name) > 63 {
		return fmt.Errorf("database name length must be between 1 and 63 characters")
	}

	// First character must be a letter or underscore
	if !regexp.MustCompile(`^[a-zA-Z_]`).MatchString(name[:1]) {
		return fmt.Errorf("database name must start with a letter or underscore")
	}

	// Remaining characters must be letters, digits, underscores, or hyphens
	if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`).MatchString(name) {
		return fmt.Errorf("database name can only contain letters, digits, underscores, and hyphens")
	}

	return nil
}

func runMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, "migrations")
}

// createDatabaseIfNotExists creates the database if it doesn't exist
func createDatabaseIfNotExists(dbURL string) error {
	// Parse the connection string to extract database name
	// Format: postgresql://user:pass@host:port/dbname?sslmode=disable
	parts := strings.Split(dbURL, "/")
	if len(parts) < 4 {
		return fmt.Errorf("invalid database URL format")
	}

	dbName := strings.Split(parts[len(parts)-1], "?")[0]

	// Validate database name before using it in SQL
	if err := isValidDatabaseName(dbName); err != nil {
		return fmt.Errorf("invalid database name '%s': %w", dbName, err)
	}

	// Create connection string to default postgres database
	defaultDBURL := strings.Replace(dbURL, "/"+dbName, "/postgres", 1)

	// Connect to default postgres database
	defaultDB, err := sql.Open("postgres", defaultDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to default database: %w", err)
	}
	defer defaultDB.Close()

	// Check if our database exists
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`
	err = defaultDB.QueryRow(query, dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !exists {
		// Create the database using parameterized query for safety
		createQuery := `CREATE DATABASE "` + dbName + `"`
		_, err = defaultDB.Exec(createQuery)
		if err != nil {
			return fmt.Errorf("failed to create database %s: %w", dbName, err)
		}
		fmt.Printf("Created database: %s\n", dbName)
	} else {
		fmt.Printf("Database already exists: %s\n", dbName)
	}

	return nil
}
