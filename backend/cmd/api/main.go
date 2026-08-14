package main

import (
	"github.com/narayan-mindfire/data-processor/backend/pkg/errors"
	"github.com/narayan-mindfire/data-processor/backend/pkg/logger"
)

func main() {
	logger.InitLogger()

	logger.Log.Info("Data Processor API Server Initialized", "port", 8080)
}
