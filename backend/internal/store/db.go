package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/narayan-mindfire/data-processor/backend/pkg/logger"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type DB struct {
	*sql.DB
}

// NewDB initializes the connection pool and runs migrations
func NewDB(host, port, user, password, dbname string) (*DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Connection Pool Best Practices
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storeDB := &DB{DB: db}

	// Run Auto-Migrations
	if err := storeDB.runMigrations(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	logger.Log.Info("PostgreSQL connected and migrations applied successfully")
	return storeDB, nil
}

// runMigrations reads the embedded SQL file and executes it
func (db *DB) runMigrations(ctx context.Context) error {
	sqlBytes, err := migrationFiles.ReadFile("migrations/001_init.sql")
	if err != nil {
		return fmt.Errorf("could not read migration file: %w", err)
	}

	_, err = db.ExecContext(ctx, string(sqlBytes))
	return err
}
