package db

import (
	"log"
	"os"
	"strings"

	"github.com/glebarez/sqlite" // <-- This is the new, correct driver
	"gorm.io/driver/postgres"
	// "gorm.io/driver/sqlite" // This old one is not used
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Init initializes and returns a GORM database connection.
// It reads the DATABASE_URL environment variable.
func Init() (*gorm.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")

	// Default to local SQLite if no URL is provided
	if dbURL == "" {
		dbURL = "sqlite://whispr.db"
		log.Println("DATABASE_URL not set, defaulting to 'sqlite://whispr.db'")
	}

	var dialector gorm.Dialector

	if strings.HasPrefix(dbURL, "postgres://") {
		// Use Postgres
		dsn := strings.TrimPrefix(dbURL, "postgres://")
		dialector = postgres.Open(dsn)
		log.Println("Connecting to PostgreSQL database...")
	} else if strings.HasPrefix(dbURL, "sqlite://") {
		// Use SQLite
		dsn := strings.TrimPrefix(dbURL, "sqlite://")
		// Use the NEW driver's Open function
		dialector = sqlite.Open(dsn) // <-- This line uses the new driver
		log.Println("Connecting to SQLite database at", dsn)
	} else {
		log.Fatalf("Invalid DATABASE_URL prefix. Must start with 'postgres://' or 'sqlite://'")
	}

	// Open the database connection
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Be quiet by default
	})

	if err != nil {
		return nil, err
	}

	// Optional: Configure connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("Database connection established.")
	return db, nil
}