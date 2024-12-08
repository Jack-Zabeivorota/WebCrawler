package main

import (
	"os"

	"main/logger"
	"main/worker"
)

func main() {
	logger.Init(os.Getenv("LOGS_DIR"))
	worker.New().Run()
}
