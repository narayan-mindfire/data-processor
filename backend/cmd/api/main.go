package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/narayan-mindfire/data-processor/backend/internal/api/repositories"
	"github.com/narayan-mindfire/data-processor/backend/internal/api/routes"
	"github.com/narayan-mindfire/data-processor/backend/internal/store"
	"github.com/narayan-mindfire/data-processor/backend/pkg/logger"
)

type Config struct {
	Port   string
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
}

func loadConfig() Config {
	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		panic("CRITICAL: DB_PASSWORD environment variable is not set!")
	}

	return Config{
		Port:   getEnv("PORT", "8080"),
		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: getEnv("DB_PORT", "5432"),
		DBUser: getEnv("DB_USER", "postgres"),
		DBPass: dbPass,
		DBName: getEnv("DB_NAME", "dataprocessor"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// @title Data Processor Pipeline API
// @version 1.0
// @description High-performance concurrent data ingestion and processing pipeline.
// @host localhost:8080
// @BasePath /
func main() {
	log := logger.New()
	cfg := loadConfig()

	db, err := store.NewDB(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)
	if err != nil {
		log.Error("failed to connect to the database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("PostgreSQL connected and migrations applied successfully")

	repo := repositories.NewPostgresJobRepository(db)
	router := routes.RegisterRoutes(repo)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	log.Info("Data Processor API Server Initialized", "port", cfg.Port)
	log.Info("Swagger Documentation available at: http://localhost:8080/api-docs/index.html")

	// Start server in a background goroutine so we don't block
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server crashed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}
}
