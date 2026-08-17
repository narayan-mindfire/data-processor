package main

import (
	"net/http"
	"os"
	"github.com/narayan-mindfire/data-processor/backend/internal/api/repositories"
	"github.com/narayan-mindfire/data-processor/backend/internal/api/routes"
	"github.com/narayan-mindfire/data-processor/backend/internal/store"
	"github.com/narayan-mindfire/data-processor/backend/pkg/logger"
)

// @title Data Processor Pipeline API
// @version 1.0
// @description High-performance concurrent data ingestion and processing pipeline.
// @host localhost:8080
// @BasePath /
func main() {
	logger.InitLogger()
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgrespassword")
	dbName := getEnv("DB_NAME", "dataprocessor")

	db, err := store.NewDB(dbHost, dbPort, dbUser, dbPass, dbName)
	if err != nil {
		logger.Log.Error("failed to connect to the database", "error", err)
		os.Exit(1)
	}

	defer db.Close()

	repo := repositories.NewPostgresJobRepository(db)
	router := routes.RegisterRoutes(repo)
	logger.Log.Info("Data Processor API Server Initialized", "port", 8080)
	logger.Log.Info("Swagger Documentation available at: http://localhost:8080/api-docs/index.html")
	
	err = http.ListenAndServe(":"+getEnv("PORT", "8080"), router)
	if err != nil {
		logger.Log.Error("server crashed", "error", err)
	}

	logger.Log.Info("Data Processor API Server Initialized and Database Connected", "port", 8080)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
