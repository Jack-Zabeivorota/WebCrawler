package main

import (
	"os"

	"main/controller"
	"main/logger"
)

func main() {
	logger.Init(os.Getenv("LOGS_DIR"))
	controller.New().Run()
}
