package main

import (
	"github.com/narayan-mindfire/data-processor/backend/pkg/logger"
	"github.com/narayan-mindfire/data-processor/backend/internal/store"
	"os"
)

func main() {
	logger.InitLogger()
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASS", "postgrespassword")
	dbName := getEnv("DB_NAME", "dataprocessor")

	db, err := store.NewDB(dbHost, dbPort, dbUser, dbPass, dbName)
	if err != nil {
		logger.Log.Error("failed to connect to the database", "error", err)
		os.Exit(1)
	}

	defer db.Close()

	logger.Log.Info("Data Processor API Server Initialized and Database Connected", "port", 8080)
}

func getEnv(key, defaultValue string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return defaultValue
}
