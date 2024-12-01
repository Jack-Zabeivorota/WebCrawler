package main

import (
	"os"

	"main/aggregator"
	"main/logger"
)

func main() {
	logger.Init(os.Getenv("LOGS_DIR"))
	aggregator.New().Run()
}
